package jmt

import (
	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func Test_CacheInsert(t *testing.T) {
	n1 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n1.Children[1] = &types.Child{
		Version: 1,
		Leaf:    true,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
	}
	n1.Children[10] = &types.Child{
		Version: 2,
		Leaf:    false,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
	}
	node1 := &NodeData{
		Node: n1,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 1,
			Path:    []byte{},
		},
	}

	n2 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n2.Children[5] = &types.Child{
		Version: 2,
		Leaf:    false,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
	}
	node2 := &NodeData{
		Node: n2,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 2,
			Path:    []byte{10},
		},
	}

	n3 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n3.Children[3] = &types.Child{
		Version: 3,
		Leaf:    true,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
	}
	node3 := &NodeData{
		Node: n3,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 3,
			Path:    []byte{10, 5},
		},
	}

	cache := NewJMTCache(1, log.NewWithModule("JMT-Test"))
	cache.BatchInsert(NodeDataList{node1, node2, node3})
	cache.PrintAll()
}

func Test_CacheDelete(t *testing.T) {
	n1 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n1.Children[1] = &types.Child{
		Version: 1,
		Leaf:    true,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
	}
	n1.Children[10] = &types.Child{
		Version: 2,
		Leaf:    false,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
	}
	node1 := &NodeData{
		Node: n1,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 1,
			Path:    []byte{},
		},
	}

	n2 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n2.Children[5] = &types.Child{
		Version: 2,
		Leaf:    false,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
	}
	node2 := &NodeData{
		Node: n2,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 2,
			Path:    []byte{10},
		},
	}

	n3 := &types.InternalNode{
		Children: [16]*types.Child{},
	}
	n3.Children[3] = &types.Child{
		Version: 3,
		Leaf:    true,
		Hash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
	}
	node3 := &NodeData{
		Node: n3,
		Nk: &types.NodeKey{
			Type:    []byte("1"),
			Version: 3,
			Path:    []byte{10, 5},
		},
	}

	cache := NewJMTCache(1, log.NewWithModule("JMT-Test"))
	cache.BatchInsert(NodeDataList{node1, node2, node3})
	cache.PrintAll()

	// delete node3

	cache.BatchDelete(NodeDataList{node3})
	cache.PrintAll()

}
