package jmt

import (
	"container/heap"
	"errors"
	"fmt"
	"time"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrorTimeout     = errors.New("wait too long when iterating trie")
	ErrorInterrupted = errors.New("interrupt iterating trie")
	ErrorNoMoreData  = errors.New("no more trie data")
)

// Iterator traverse whole jmt trie
type Iterator struct {
	rootHash common.Hash
	backend  storage.Storage

	bufferC chan *RawNode // use buffer to balance between read and write
	errC    chan error
	stopC   chan struct{}
	timeout time.Duration

	nodeKeyHeap *NodeKeyHeap // max heap, store NodeKeys that are waiting to be visited
}

func NewIterator(rootHash common.Hash, backend storage.Storage, bufSize int, timeout time.Duration) *Iterator {
	nodeKeyHeap := &NodeKeyHeap{}
	heap.Init(nodeKeyHeap)
	return &Iterator{
		rootHash:    rootHash,
		backend:     backend,
		bufferC:     make(chan *RawNode, bufSize),
		errC:        make(chan error, 1),
		stopC:       make(chan struct{}),
		timeout:     timeout,
		nodeKeyHeap: nodeKeyHeap,
	}
}

func (it *Iterator) Iterate() {
	defer func() {
		close(it.stopC)
		close(it.bufferC)
		close(it.errC)
	}()

	// initialize trie
	rawRootNodeKey := it.backend.Get(it.rootHash[:])
	if rawRootNodeKey == nil {
		it.errC <- ErrorNotFound
		return
	}
	rootNodeKey := decodeNodeKey(rawRootNodeKey)
	heap.Push(it.nodeKeyHeap, rootNodeKey)

	for it.nodeKeyHeap.Len() != 0 {
		// pop current node from heap
		currentNodeKey := heap.Pop(it.nodeKeyHeap).(*NodeKey)

		// get current node from kv
		var currentNode types.Node
		nk := currentNodeKey.encode()
		currentNodeBlob := it.backend.Get(nk)
		currentNode, err := types.UnmarshalJMTNodeFromPb(currentNodeBlob)
		if err != nil {
			it.errC <- err
			return
		}

		var leafKey, leafValue []byte
		n, ok := currentNode.(*types.InternalNode)
		if !ok {
			leaf, _ := currentNode.(*types.LeafNode)
			leafValue = leaf.Val
			leafKey = leaf.Key
		}

		select {
		case <-it.stopC:
			it.errC <- ErrorInterrupted
			return
		case <-time.After(it.timeout):
			it.errC <- ErrorTimeout
			return
		case it.bufferC <- &RawNode{
			RawKey:    nk,
			RawValue:  currentNodeBlob,
			LeafKey:   leafKey,
			LeafValue: leafValue,
		}:
		}

		// continue if current node is leaf node
		if !ok {
			continue
		}

		var hex byte
		for hex = 0; hex < 16; hex++ {
			if n.Children[hex] == nil {
				continue
			}
			child := n.Children[hex]
			childNodeKey := &NodeKey{
				Version: child.Version,
				Type:    rootNodeKey.Type,
				Path:    make([]byte, len(currentNodeKey.Path)),
			}
			copy(childNodeKey.Path, currentNodeKey.Path)
			childNodeKey.Path = append(childNodeKey.Path, hex)

			// push child node's nodeKey to heap
			heap.Push(it.nodeKeyHeap, childNodeKey)
		}
	}
}

func (it *Iterator) IterateLeaf() {
	defer func() {
		close(it.stopC)
		close(it.bufferC)
		close(it.errC)
	}()

	// initialize trie
	rawRootNodeKey := it.backend.Get(it.rootHash[:])
	if rawRootNodeKey == nil {
		it.errC <- ErrorNotFound
		return
	}
	rootNodeKey := decodeNodeKey(rawRootNodeKey)
	heap.Push(it.nodeKeyHeap, rootNodeKey)

	for it.nodeKeyHeap.Len() != 0 {
		// pop current node from heap
		currentNodeKey := heap.Pop(it.nodeKeyHeap).(*NodeKey)

		// get current node from kv
		var currentNode types.Node
		nk := currentNodeKey.encode()
		currentNodeBlob := it.backend.Get(nk)
		currentNode, err := types.UnmarshalJMTNodeFromPb(currentNodeBlob)
		if err != nil {
			it.errC <- err
			return
		}

		var leafKey, leafValue []byte
		n, ok := currentNode.(*types.InternalNode)

		// current node is leaf node
		if !ok {
			leaf, _ := currentNode.(*types.LeafNode)
			leafValue = leaf.Val
			leafKey = leaf.Key
			select {
			case <-it.stopC:
				it.errC <- ErrorInterrupted
				return
			case <-time.After(it.timeout):
				it.errC <- ErrorTimeout
				return
			case it.bufferC <- &RawNode{
				LeafKey:   leafKey,
				LeafValue: leafValue,
			}:
			}
			continue
		}

		// current node is internal node
		var hex byte
		for hex = 0; hex < 16; hex++ {
			if n.Children[hex] == nil {
				continue
			}
			child := n.Children[hex]
			childNodeKey := &NodeKey{
				Version: child.Version,
				Type:    rootNodeKey.Type,
				Path:    make([]byte, len(currentNodeKey.Path)),
			}
			copy(childNodeKey.Path, currentNodeKey.Path)
			childNodeKey.Path = append(childNodeKey.Path, hex)

			// push child node's nodeKey to heap
			heap.Push(it.nodeKeyHeap, childNodeKey)
		}
	}
}

func (it *Iterator) Stop() {
	it.stopC <- struct{}{}
}

func (it *Iterator) Next() (*RawNode, error) {
	for {
		select {
		case err, ok := <-it.errC:
			if ok {
				return nil, err
			}
		case node, ok := <-it.bufferC:
			if ok {
				return node, nil
			}
			return nil, ErrorNoMoreData
		}
	}
}

type RawNode struct {
	RawKey    []byte // physical key of trie node in KV storage
	RawValue  []byte // physical value of trie node in KV storage
	LeafKey   []byte // non-empty iff current node is a leaf node, represents logical key of leaf
	LeafValue []byte // non-empty iff current node is a leaf node, represents logical value of leaf
}

// just for debug
func (n *RawNode) String() string {
	nk := decodeNodeKey(n.RawKey)
	node, _ := types.UnmarshalJMTNodeFromPb(n.RawValue)
	return fmt.Sprintf("[RawNode]: nodeKey={%v}, nodeValue={%v}", nk.print(), node.String())
}
