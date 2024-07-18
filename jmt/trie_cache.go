package jmt

import (
	"container/heap"
	"encoding/binary"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	trieCacheHitCount  atomic.Uint32
	trieCacheMissCount atomic.Uint32

	trieCacheHitCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jmt",
		Name:      "trie_cache_hit_counter_per_block",
		Help:      "The total number of trie cache hit per block",
	})

	trieCacheMissCounterPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jmt",
		Name:      "trie_cache_miss_counter_per_block",
		Help:      "The total number of trie cache miss per block",
	})

	trieCacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "axiom_ledger",
		Subsystem: "jmt",
		Name:      "trie_cache_active_size",
		Help:      "The total mem space of active trie cache (MB)",
	})
)

func init() {
	prometheus.MustRegister(trieCacheHitCounterPerBlock)
	prometheus.MustRegister(trieCacheMissCounterPerBlock)
	prometheus.MustRegister(trieCacheSize)
}

func (c *JMTCache) ExportTrieCacheMetrics() {
	trieCacheHitCounterPerBlock.Set(float64(trieCacheHitCount.Load()))
	trieCacheMissCounterPerBlock.Set(float64(trieCacheMissCount.Load()))
	trieCacheSize.Set(float64((childSlotSize*c.childBuffer.usedSize + headerSlotSize*c.headerBuffer.usedSize) / 1024 / 1024))
}

func (c *JMTCache) ResetTrieCacheMetrics() {
	trieCacheHitCount.Store(0)
	trieCacheMissCount.Store(0)

}

const (
	headerSlotSize   = 68
	childSlotSize    = 45
	bufferOffsetSize = 4
)

type (
	JMTCache struct {
		headerBuffer  *HeaderBuffer
		childBuffer   *ChildBuffer
		nodeDepthHeap *NodeDepthHeap

		rootOffsetMap map[string]BufferOffset // identifier of available trie -> offset of its root in headerBuffer

		// todo lru暂存最近数据

		size int64 // total cache size in byte

		lock   sync.RWMutex
		logger logrus.FieldLogger
	}
)

type (
	// CachedHeader 68B
	CachedHeader struct {
		Childs             [types.TrieDegree]BufferOffset // 16个孩子节点在buffer中的偏移量，0表示孩子不存在
		ParentHeaderOffset BufferOffset                   // 当前节点的Parent节点在buffer中的偏移量(下标)
	}

	// CachedChild 45B
	CachedChild struct {
		Hash              [32]byte
		Version           [8]byte
		Leaf              byte
		ChildHeaderOffset BufferOffset // 孩子节点指向的下一层中间节点头部的下标
	}
)

type (
	HeaderBuffer struct {
		buffer    []HeaderSlot  // each element in buffer is a header
		bitMap    *BufferBitmap // free slot in bitmap is marked as 0
		current   BufferOffset  // offset of the latest used slot
		usedSize  uint32        // number of used slots
		totalSize uint32
	}

	ChildBuffer struct {
		buffer    []ChildSlot   // each element in buffer is a child
		bitMap    *BufferBitmap // free slot in bitmap is marked as 0
		current   BufferOffset  // offset of the latest used slot
		usedSize  uint32        // number of used slots
		totalSize uint32
	}

	HeaderSlot   = [headerSlotSize]byte
	ChildSlot    = [childSlotSize]byte
	BufferOffset = uint32

	// 堆元素，固定5B=1B的深度+4B的Offset
	NodeDepthHeapItem = [5]byte

	NodeDepthHeap struct {
		heap      []NodeDepthHeapItem       // 树深度大根堆，记录缓存中各个节点在树的深度信息
		headIndex map[NodeDepthHeapItem]int // 记录缓存中存活的树节点的深度信息在堆中的索引，便于删除
	}

	BufferBitmap struct {
		bits []uint64
		size BufferOffset
	}

	// todo consider simplify this data structure
	NodeData struct {
		Nk   *types.NodeKey
		Node types.Node
	}

	NodeDataList []*NodeData
)

var zeroOffset = []byte{0, 0, 0, 0}

// todo check megabytesLimit range overflow
// todo rebuild cache after restart node
func NewJMTCache(megabytesLimit int, logger logrus.FieldLogger) *JMTCache {
	if megabytesLimit <= 0 {
		megabytesLimit = 128
	}

	nodeDepthHeap := &NodeDepthHeap{}
	heap.Init(nodeDepthHeap)

	headerNum := BufferOffset(megabytesLimit * 1024 * 1024 / 10 / headerSlotSize)
	childNum := BufferOffset(megabytesLimit * 1024 * 1024 / 10 * 9 / childSlotSize)

	cache := &JMTCache{
		headerBuffer: &HeaderBuffer{
			buffer:    make([][headerSlotSize]byte, headerNum),
			bitMap:    NewBufferBitmap(headerNum),
			totalSize: headerNum,
			usedSize:  0,
		},
		childBuffer: &ChildBuffer{
			buffer:    make([][childSlotSize]byte, childNum),
			bitMap:    NewBufferBitmap(childNum),
			totalSize: childNum,
			usedSize:  0,
		},
		nodeDepthHeap: nodeDepthHeap,
		rootOffsetMap: make(map[string]BufferOffset),
		logger:        logger,
	}
	return cache
}

func (c *JMTCache) Get(nk *types.NodeKey) (res types.Node, exist bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	defer func() {
		if exist {
			trieCacheHitCount.Add(1)
		} else {
			trieCacheMissCount.Add(1)
		}
	}()

	nextHeaderOffset, ok := c.rootOffsetMap[string(nk.Type)]
	if !ok {
		return nil, false
	}

	currentHeader := c.headerBuffer.buffer[nextHeaderOffset]
	for currentLoc := 0; currentLoc < len(nk.Path); currentLoc++ {
		nextBranch := nk.Path[currentLoc]
		nextChildOffset := binary.BigEndian.Uint32(currentHeader[nextBranch*bufferOffsetSize : (nextBranch+1)*bufferOffsetSize])
		nextChild := c.childBuffer.buffer[nextChildOffset]
		nextHeaderOffset = binary.BigEndian.Uint32(nextChild[childSlotSize-bufferOffsetSize:])
		if nextHeaderOffset == 0 {
			return nil, false
		}
		currentHeader = c.headerBuffer.buffer[nextHeaderOffset]
	}

	// construct node instance
	n := types.NewInternalNode()
	for i := 0; i < types.TrieDegree; i++ {
		childOffset := binary.BigEndian.Uint32(currentHeader[i*bufferOffsetSize : (i+1)*bufferOffsetSize])
		if childOffset == 0 {
			n.Children[i] = nil
		} else {
			child := c.childBuffer.buffer[childOffset]
			n.Children[i] = &types.Child{
				Hash:    common.BytesToHash(child[:32]),
				Version: binary.BigEndian.Uint64(child[32:40]),
				Leaf:    child[40] == 1,
			}
		}
	}

	return n, true
}

func (c *JMTCache) Has(nk *types.NodeKey) bool {
	return false
}

func (c *JMTCache) Update(insertNodes, delNodes NodeDataList) {
	realDelNodes := make(NodeDataList, 0)
	insertSet := make(map[string]map[string]struct{})
	for _, n := range insertNodes {
		if _, ok := insertSet[string(n.Nk.Type)]; !ok {
			insertSet[string(n.Nk.Type)] = make(map[string]struct{})
		}
		insertSet[string(n.Nk.Type)][string(n.Nk.Path)] = struct{}{}
	}
	for _, n := range delNodes {
		if _, ok := insertSet[string(n.Nk.Type)]; !ok {
			realDelNodes = append(realDelNodes, n)
			continue
		}
		if _, ok := insertSet[string(n.Nk.Type)][string(n.Nk.Path)]; !ok {
			realDelNodes = append(realDelNodes, n)
		}
	}
	current := time.Now()
	c.delete(realDelNodes)
	delTime := time.Since(current)

	current = time.Now()
	c.insert(insertNodes)
	c.logger.Infof("[JMTCache-Update] kv del: %v, real del: %v, time: %v, insert: %v, time: %v",
		len(delNodes), len(realDelNodes), delTime, len(insertNodes), time.Since(current))
}

// todo delete to empty case
func (c *JMTCache) delete(nodes NodeDataList) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// sort trie nodes from top to bottom
	sort.Sort(nodes)

	headerBuffer := c.headerBuffer.buffer
	childBuffer := c.childBuffer.buffer
	for i := len(nodes) - 1; i >= 0; i-- {
		currentHeaderOffset, ok := c.rootOffsetMap[string(nodes[i].Nk.Type)]
		if !ok {
			continue
		}
		currentSlot := byte(0) // current header's location in its parent header
		for depth := 0; depth < len(nodes[i].Nk.Path); depth++ {
			currentSlot = nodes[i].Nk.Path[depth]
			childOffset := binary.BigEndian.Uint32(headerBuffer[currentHeaderOffset][currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
			currentHeaderOffset = binary.BigEndian.Uint32(childBuffer[childOffset][childSlotSize-bufferOffsetSize:])
			if currentHeaderOffset == 0 {
				break
			}
		}
		if currentHeaderOffset == 0 {
			continue
		}
		// mark all current header's child slot as free
		for childSlot := byte(0); childSlot < types.TrieDegree; childSlot++ {
			childOffset := binary.BigEndian.Uint32(headerBuffer[currentHeaderOffset][childSlot*bufferOffsetSize : (childSlot+1)*bufferOffsetSize])
			c.childBuffer.resetSlot(childOffset)
		}
		// dereference current header in its parent header
		parentHeaderOffset := binary.BigEndian.Uint32(headerBuffer[currentHeaderOffset][headerSlotSize-bufferOffsetSize:])
		childOffset := binary.BigEndian.Uint32(headerBuffer[parentHeaderOffset][currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
		binary.BigEndian.PutUint32(childBuffer[childOffset][childSlotSize-bufferOffsetSize:], 0)
		// mark current header slot as free
		c.headerBuffer.resetSlot(currentHeaderOffset)
	}
}

func (c *JMTCache) insert(nodes NodeDataList) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// sort trie nodes from top to bottom
	sort.Sort(nodes)

	headerBuffer := c.headerBuffer.buffer
	childBuffer := c.childBuffer.buffer
	for i := range nodes {
		var parentHeaderOffset BufferOffset
		var currentSlot, begin, end byte
		currentHeaderOffset, rootExist := c.rootOffsetMap[string(nodes[i].Nk.Type)]
		for depth := 0; depth < len(nodes[i].Nk.Path); depth++ {
			currentSlot = nodes[i].Nk.Path[depth]
			begin, end = currentSlot*bufferOffsetSize, (currentSlot+1)*bufferOffsetSize
			childOffset := binary.BigEndian.Uint32(headerBuffer[currentHeaderOffset][begin:end])
			parentHeaderOffset = currentHeaderOffset
			currentHeaderOffset = binary.BigEndian.Uint32(childBuffer[childOffset][childSlotSize-bufferOffsetSize:])
			if currentHeaderOffset == 0 {
				break
			}
		}

		// current header is a new header
		if currentHeaderOffset == 0 {
			// reference parent header in current header
			currentHeaderOffset = c.headerBuffer.allocSlot()
			binary.BigEndian.PutUint32(headerBuffer[currentHeaderOffset][headerSlotSize-bufferOffsetSize:], parentHeaderOffset)

			// reference current header in its parent header
			if parentHeaderOffset != 0 {
				parentChildOffset := binary.BigEndian.Uint32(headerBuffer[parentHeaderOffset][begin:end])
				binary.BigEndian.PutUint32(childBuffer[parentChildOffset][childSlotSize-bufferOffsetSize:], currentHeaderOffset)
			}
		}
		if !rootExist {
			c.rootOffsetMap[string(nodes[i].Nk.Type)] = currentHeaderOffset
		}

		// update all current header's child offset
		n := nodes[i].Node.(*types.InternalNode)
		for childSlot := byte(0); childSlot < types.TrieDegree; childSlot++ {
			childBegin, childEnd := childSlot*bufferOffsetSize, (childSlot+1)*bufferOffsetSize
			childOffset := binary.BigEndian.Uint32(headerBuffer[currentHeaderOffset][childBegin:childEnd])
			if n.Children[childSlot] == nil {
				// delete current child slot
				if childOffset != 0 {
					c.childBuffer.resetSlot(childOffset)
					copy(headerBuffer[currentHeaderOffset][childBegin:childEnd], zeroOffset)
				}
				continue
			}

			// current child is new child, reference it in current header
			if childOffset == 0 {
				childOffset = c.childBuffer.allocSlot()
				binary.BigEndian.PutUint32(headerBuffer[currentHeaderOffset][childBegin:childEnd], childOffset)
			}
			// build child slot with Hash, Version and Leaf field, but leave Next field empty
			copy(childBuffer[childOffset][:], n.Children[childSlot].Hash[:])
			binary.BigEndian.PutUint64(childBuffer[childOffset][32:40], n.Children[childSlot].Version)
			leaf := byte(0)
			if n.Children[childSlot].Leaf {
				leaf = 1
			}
			childBuffer[childOffset][40] = leaf
		}
	}
}

func (c *JMTCache) Enable() bool {
	return c != nil
}

func (c *JMTCache) PrintAll() {
	for k, v := range c.rootOffsetMap {
		trieRes := c.printCacheTrie(v)
		addr := common.BytesToAddress([]byte(k))
		c.logger.Infof("[JMTCache-PrintAll] trie root: %v, trie content:\n%v", addr, trieRes)
	}
}

func (c *JMTCache) printCacheTrie(rootHeaderOffset BufferOffset) string {
	res := strings.Builder{}
	queue := make([]BufferOffset, 1)
	queue[0] = rootHeaderOffset
	for len(queue) != 0 {
		currentHeader := c.headerBuffer.buffer[queue[0]]
		res.WriteString("{" + strconv.Itoa(int(queue[0])) + ": ")
		for branch := 0; branch < types.TrieDegree; branch++ {
			nextChildOffset := binary.BigEndian.Uint32(currentHeader[branch*bufferOffsetSize : (branch+1)*bufferOffsetSize])
			if nextChildOffset == 0 {
				res.WriteString("" + strconv.Itoa(branch) + ", ")
				continue
			}
			res.WriteString("<" + strconv.Itoa(branch) + ": ")
			nextChild := c.childBuffer.buffer[nextChildOffset]
			nextHeaderOffset := binary.BigEndian.Uint32(nextChild[childSlotSize-bufferOffsetSize:])
			res.WriteString("H[")
			res.WriteString(common.BytesToHash(nextChild[:32]).String()[62:66])
			res.WriteString("], V[")
			res.WriteString(strconv.Itoa(int(binary.BigEndian.Uint64(nextChild[32:40]))))
			res.WriteString("], L[")
			if nextChild[40] == 1 {
				res.WriteString("1")
			} else {
				res.WriteString("0")
			}
			res.WriteString("], Next[")
			res.WriteString(strconv.Itoa(int(nextHeaderOffset)))
			res.WriteString("]>, ")
			if nextHeaderOffset != 0 {
				queue = append(queue, nextHeaderOffset)
			}
		}
		res.WriteString("<Parent[" + strconv.Itoa(int(binary.BigEndian.Uint32(currentHeader[64:68]))))
		res.WriteString("]>}\n")
		queue = queue[1:]
	}
	return res.String()
}

func (cb *ChildBuffer) resetSlot(offset BufferOffset) {
	cb.bitMap.Remove(offset)
	cb.buffer[offset] = ChildSlot{}
	cb.usedSize--
}

// todo gc
func (cb *ChildBuffer) allocSlot() BufferOffset {
	if cb.usedSize == cb.totalSize-1 {
		panic("child buffer is full")
	}

	for i := cb.current + 1; i < cb.bitMap.size; i++ {
		if cb.bitMap.bits[i/64]&(1<<(i%64)) == 0 {
			cb.bitMap.bits[i/64] |= 1 << (i % 64)
			cb.usedSize++
			cb.current = i
			return i
		}
		if cb.current == cb.totalSize-1 {
			cb.current = 0
			i = 1
		}
	}
	panic("child buffer error")
}

func (hb *HeaderBuffer) resetSlot(offset BufferOffset) {
	hb.bitMap.Remove(offset)
	hb.buffer[offset] = HeaderSlot{}
	hb.usedSize--
}

// todo gc
func (hb *HeaderBuffer) allocSlot() BufferOffset {
	if hb.usedSize == hb.totalSize-1 {
		panic("header buffer is full")
	}

	for i := hb.current + 1; i < hb.bitMap.size; i++ {
		if hb.bitMap.bits[i/64]&(1<<(i%64)) == 0 {
			hb.bitMap.bits[i/64] |= 1 << (i % 64)
			hb.usedSize++
			hb.current = i
			return i
		}
		if hb.current == hb.totalSize-1 {
			hb.current = 0
			i = 1
		}
	}
	panic("header buffer error")
}

func (h *NodeDepthHeap) Len() int {
	return len(h.heap)
}

func (h *NodeDepthHeap) Less(i, j int) bool {
	return h.heap[i][1] > h.heap[j][1]
}

func (h *NodeDepthHeap) Swap(i, j int) {
	h.heap[i], h.heap[j] = h.heap[j], h.heap[i]
}

func (h *NodeDepthHeap) Push(x any) {
	h.heap = append(h.heap, x.(NodeDepthHeapItem))
}

func (h *NodeDepthHeap) Pop() any {
	old := h.heap
	n := len(old)
	x := old[n-1]
	h.heap = old[0 : n-1]
	return x
}

func NewBufferBitmap(size BufferOffset) *BufferBitmap {
	return &BufferBitmap{
		bits: make([]uint64, (size+63)/64),
		size: size,
	}
}

//func (b *BufferBitmap) Add(num BufferOffset) {
//	if num >= b.size {
//		return
//	}
//	b.bits[num/64] |= 1 << (num % 64)
//}

func (b *BufferBitmap) Remove(num BufferOffset) {
	if num >= b.size {
		return
	}
	b.bits[num/64] &^= 1 << (num % 64)
}

func (l NodeDataList) Len() int {
	return len(l)
}

func (l NodeDataList) Less(i, j int) bool {
	return len(l[i].Nk.Path) < len(l[j].Nk.Path)
}

func (l NodeDataList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
