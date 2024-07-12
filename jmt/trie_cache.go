package jmt

import (
	"container/heap"
	"encoding/binary"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"sort"
	"sync"
)

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
		buffer []HeaderSlot  // 每个槽就是一个68B的中间节点头部
		bitMap *BufferBitmap // 记录哪些槽是空闲的
	}

	ChildBuffer struct {
		buffer []ChildSlot   // 每个槽就是一个45B的孩子
		bitMap *BufferBitmap // 记录哪些槽是空闲的
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

	NodeData struct {
		Nk   *types.NodeKey
		Node types.Node
	}

	NodeDataList []*NodeData
)

var zeroOffset = []byte{0, 0, 0, 0}

// todo check megabytesLimit range overflow
func NewJMTCache(gigabytesLimit int64, logger logrus.FieldLogger) *JMTCache {
	if gigabytesLimit <= 0 {
		gigabytesLimit = 1
	}

	nodeDepthHeap := &NodeDepthHeap{}
	heap.Init(nodeDepthHeap)

	headerNum := BufferOffset(gigabytesLimit * 1024 * 1024 * 1024 / 10 / headerSlotSize)
	childNum := BufferOffset(gigabytesLimit * 1024 * 1024 * 1024 / 10 * 9 / childSlotSize)

	cache := &JMTCache{
		headerBuffer: &HeaderBuffer{
			buffer: make([][headerSlotSize]byte, headerNum),
			bitMap: NewBufferBitmap(headerNum),
		},
		childBuffer: &ChildBuffer{
			buffer: make([][childSlotSize]byte, childNum),
			bitMap: NewBufferBitmap(headerNum),
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
		c.logger.Infof("[JMTCache-Get] res=%v,exist=%v", res.String(), exist)
	}()

	nextHeaderIndex, ok := c.rootOffsetMap[string(nk.Type)]
	if !ok {
		return nil, false
	}

	currentHeader := c.headerBuffer.buffer[nextHeaderIndex]
	for currentLoc := 0; currentLoc < len(nk.Path); currentLoc++ {
		nextBranch := nk.Path[currentLoc]
		nextChildOffset := binary.BigEndian.Uint32(currentHeader[nextBranch*bufferOffsetSize : (nextBranch+1)*bufferOffsetSize])
		nextChild := c.childBuffer.buffer[nextChildOffset]
		nextHeaderIndex = binary.BigEndian.Uint32(nextChild[childSlotSize-bufferOffsetSize:])
		if nextHeaderIndex == 0 {
			return nil, false
		}
		currentHeader = c.headerBuffer.buffer[nextHeaderIndex]
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

func (c *JMTCache) BatchDelete(nodes NodeDataList) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// sort trie nodes from top to bottom
	sort.Sort(nodes)

	for i := len(nodes) - 1; i >= 0; i-- {
		currentHeaderOffset, ok := c.rootOffsetMap[string(nodes[i].Nk.Type)]
		if !ok {
			continue
		}
		currentHeader := c.headerBuffer.buffer[currentHeaderOffset]
		currentSlot := byte(0) // current header's location in its parent header
		for depth := 0; depth < len(nodes[i].Nk.Path); depth++ {
			currentSlot = nodes[i].Nk.Path[depth]
			childOffset := binary.BigEndian.Uint32(currentHeader[currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
			child := c.childBuffer.buffer[childOffset]
			currentHeaderOffset = binary.BigEndian.Uint32(child[childSlotSize-bufferOffsetSize:])
			if currentHeaderOffset == 0 {
				break
			}
			currentHeader = c.headerBuffer.buffer[currentHeaderOffset]
		}
		if currentHeaderOffset == 0 {
			continue
		}
		// mark current header slot as free
		c.headerBuffer.resetSlot(currentHeaderOffset)
		// mark all current header's child slot as free
		for childSlot := byte(0); childSlot < types.TrieDegree; childSlot++ {
			childOffset := binary.BigEndian.Uint32(currentHeader[childSlot*bufferOffsetSize : (childSlot+1)*bufferOffsetSize])
			c.childBuffer.resetSlot(childOffset)
		}
		// dereference current header in its parent header
		parentHeaderOffset := binary.BigEndian.Uint32(currentHeader[(types.TrieDegree-1)*bufferOffsetSize:])
		parentHeader := c.headerBuffer.buffer[parentHeaderOffset]
		childOffset := binary.BigEndian.Uint32(parentHeader[currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
		c.childBuffer.resetSlot(childOffset)
		copy(parentHeader[(types.TrieDegree-1)*bufferOffsetSize:], zeroOffset)
	}
}

func (c *JMTCache) BatchInsert(nodes NodeDataList) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// sort trie nodes from top to bottom
	sort.Sort(nodes)

	for i := range nodes {
		var parentHeaderOffset BufferOffset
		currentHeaderOffset, rootExist := c.rootOffsetMap[string(nodes[i].Nk.Type)]
		currentHeader := c.headerBuffer.buffer[currentHeaderOffset]
		currentSlot := byte(0) // current header's location in its parent header
		for depth := 0; depth < len(nodes[i].Nk.Path); depth++ {
			currentSlot = nodes[i].Nk.Path[depth]
			childOffset := binary.BigEndian.Uint32(currentHeader[currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
			child := c.childBuffer.buffer[childOffset]
			parentHeaderOffset = currentHeaderOffset
			currentHeaderOffset = binary.BigEndian.Uint32(child[childSlotSize-bufferOffsetSize:])
			if currentHeaderOffset == 0 {
				break
			}
			currentHeader = c.headerBuffer.buffer[currentHeaderOffset]
		}

		// current header is a new header
		if currentHeaderOffset == 0 {
			// reference parent header in current header
			currentHeaderOffset = c.headerBuffer.findFreeOffset()
			currentHeader = c.headerBuffer.buffer[currentHeaderOffset]
			binary.BigEndian.PutUint32(currentHeader[(types.TrieDegree-1)*bufferOffsetSize:], parentHeaderOffset)

			// reference current header in its parent header
			if parentHeaderOffset != 0 {
				parentHeader := c.headerBuffer.buffer[parentHeaderOffset]
				parentChildOffset := binary.BigEndian.Uint32(parentHeader[currentSlot*bufferOffsetSize : (currentSlot+1)*bufferOffsetSize])
				binary.BigEndian.PutUint32(c.childBuffer.buffer[parentChildOffset][41:45], currentHeaderOffset)
			}
		}
		if !rootExist {
			c.rootOffsetMap[string(nodes[i].Nk.Type)] = currentHeaderOffset
		}

		// update all current header's child offset
		n := nodes[i].Node.(*types.InternalNode)
		for childSlot := byte(0); childSlot < types.TrieDegree; childSlot++ {
			childOffset := binary.BigEndian.Uint32(currentHeader[childSlot*bufferOffsetSize : (childSlot+1)*bufferOffsetSize])
			if n.Children[childSlot] == nil {
				// delete current child slot
				if childOffset != 0 {
					c.childBuffer.resetSlot(childOffset)
					copy(currentHeader[(types.TrieDegree-1)*bufferOffsetSize:], zeroOffset)
				}
				continue
			}

			// current child is new child
			if childOffset == 0 {
				childOffset = c.childBuffer.findFreeOffset()
				binary.BigEndian.PutUint32(currentHeader[childSlot*bufferOffsetSize:(childSlot+1)*bufferOffsetSize], childOffset)
			}
			// build child slot with Hash, Version and Leaf
			copy(c.childBuffer.buffer[childOffset][:], n.Children[childSlot].Hash[:])
			binary.BigEndian.PutUint64(c.childBuffer.buffer[childOffset][32:40], n.Children[childSlot].Version)
			leaf := byte(0)
			if n.Children[childSlot].Leaf {
				leaf = 1
			}
			c.childBuffer.buffer[childOffset][40] = leaf
		}
	}
}

func (c *JMTCache) Enable() bool {
	return c != nil
}

func (cb *ChildBuffer) resetSlot(offset BufferOffset) {
	cb.bitMap.Remove(offset)
	cb.buffer[offset] = ChildSlot{}
}

func (cb *ChildBuffer) findFreeOffset() BufferOffset {
	for i := BufferOffset(0); i < cb.bitMap.size; i++ {
		if cb.bitMap.bits[i/64]&(1<<(i%64)) == 0 {
			return i
		}
	}
	panic("child buffer is full")
}

func (hb *HeaderBuffer) resetSlot(offset BufferOffset) {
	hb.bitMap.Remove(offset)
	hb.buffer[offset] = HeaderSlot{}
}

func (hb *HeaderBuffer) findFreeOffset() BufferOffset {
	for i := BufferOffset(0); i < hb.bitMap.size; i++ {
		if hb.bitMap.bits[i/64]&(1<<(i%64)) == 0 {
			return i
		}
	}
	panic("header buffer is full")
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

func (b *BufferBitmap) Add(num BufferOffset) {
	if num >= b.size {
		return
	}
	b.bits[num/64] |= 1 << (num % 64)
}

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
