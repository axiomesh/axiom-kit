package types

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestInternalNode_Marshal(t *testing.T) {
	children := [16]*Child{}
	children[1] = &Child{
		Version: 1,
		Leaf:    true,
		Hash:    common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef26"),
	}
	children[10] = &Child{
		Version: 2,
		Leaf:    false,
		Hash:    common.HexToHash("0x96c9d6081a6c4afcdace50362be2cdc72dff324ace17b6226284f6982d949e85"),
	}

	n1 := &InternalNode{
		Children: children,
	}

	blob := n1.EncodePb()
	assert.NotNil(t, blob)
	node, err := UnmarshalJMTNodeFromPb(blob)
	assert.Nil(t, err)
	assert.True(t, equalInternalNode(n1, node.(*InternalNode)))

	n1.Children[10] = nil
	blob2 := n1.EncodePb()
	assert.NotNil(t, blob2)
	node1, err := UnmarshalJMTNodeFromPb(blob2)
	assert.False(t, equalInternalNode(node1.(*InternalNode), node.(*InternalNode)))

	assert.True(t, len(blob) > len(blob2))
}

func TestLeafNode_Marshal(t *testing.T) {
	leaf1 := &LeafNode{
		Key:  []byte{10, 1, 9, 3},
		Val:  []byte("value"),
		Hash: common.HexToHash("0x4d5e855f8fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef26"),
	}

	blob := leaf1.EncodePb()

	node, err := UnmarshalJMTNodeFromPb(blob)
	assert.Nil(t, err)

	assert.True(t, equalLeafNode(leaf1, node.(*LeafNode)))

}

func equalInternalNode(n1, n2 *InternalNode) bool {
	if n1 == nil && n2 == nil {
		return true
	}
	if n1 == nil && n2 != nil || n1 != nil && n2 == nil {
		return false
	}

	for i := 0; i < 16; i++ {
		if n1.Children[i] == nil && n2.Children[i] == nil {
			continue
		}
		if n1.Children[i] == nil && n2.Children[i] != nil ||
			n1.Children[i] != nil && n2.Children[i] == nil {
			return false
		}
		if n1.Children[i].Leaf != n2.Children[i].Leaf ||
			n1.Children[i].Version != n2.Children[i].Version ||
			n1.Children[i].Hash != n2.Children[i].Hash {
			return false
		}
	}

	return true
}

func equalLeafNode(n1, n2 *LeafNode) bool {
	if n1 == nil && n2 == nil {
		return true
	}
	if n1 == nil && n2 != nil || n1 != nil && n2 == nil {
		return false
	}

	if bytes.Equal(n1.Key, n2.Key) && bytes.Equal(n1.Val, n2.Val) && n1.Hash == n2.Hash {
		return true
	}

	return false
}
