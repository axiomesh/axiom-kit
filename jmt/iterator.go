package jmt

import (
	"container/heap"
	"errors"
	"fmt"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

var (
	ErrorTimeout     = errors.New("wait too long when iterating trie")
	ErrorInterrupted = errors.New("interrupt iterating trie")
)

// Iterator traverse whole jmt trie
type Iterator struct {
	rootHash common.Hash
	backend  storage.Storage

	BufferC chan *RawNode // use buffer to balance between read and write
	ErrC    chan error
	StopC   chan struct{}
	timeout time.Duration

	nodeKeyHeap *NodeKeyHeap // max heap, store NodeKeys that are waiting to be visited
}

func NewIterator(rootHash common.Hash, backend storage.Storage, bufSize int, timeout time.Duration) *Iterator {
	nodeKeyHeap := &NodeKeyHeap{}
	heap.Init(nodeKeyHeap)
	return &Iterator{
		rootHash:    rootHash,
		backend:     backend,
		BufferC:     make(chan *RawNode, bufSize),
		ErrC:        make(chan error, 1),
		StopC:       make(chan struct{}),
		timeout:     timeout,
		nodeKeyHeap: nodeKeyHeap,
	}
}

func (it *Iterator) Iterate() {
	defer func() {
		close(it.StopC)
		close(it.BufferC)
		close(it.ErrC)
	}()

	// initialize trie
	rawRootNodeKey := it.backend.Get(it.rootHash[:])
	if rawRootNodeKey == nil {
		it.ErrC <- ErrorNotFound
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
		currentNode, err := types.UnmarshalJMTNode(currentNodeBlob)
		if err != nil {
			it.ErrC <- err
			return
		}

		var leafContent []byte
		n, ok := currentNode.(*types.InternalNode)
		if !ok {
			leaf, _ := currentNode.(*types.LeafNode)
			leafContent = leaf.Val
		}

		select {
		case <-it.StopC:
			it.ErrC <- ErrorInterrupted
			return
		case <-time.After(it.timeout):
			it.ErrC <- ErrorTimeout
			return
		case it.BufferC <- &RawNode{
			RawKey:      nk,
			RawValue:    currentNodeBlob,
			LeafContent: leafContent,
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

func (it *Iterator) Stop() {
	it.StopC <- struct{}{}
}

// RawNode represents trie node in physical storage
type RawNode struct {
	RawKey      []byte
	RawValue    []byte
	LeafContent []byte // non-empty iff current node is a leaf node
}

// just for debug
func (n *RawNode) print() string {
	nk := decodeNodeKey(n.RawKey)
	node, _ := types.UnmarshalJMTNode(n.RawValue)
	return fmt.Sprintf("[RawNode]: nodeKey={%v}, nodeValue={%v}", nk.print(), node.Print())
}
