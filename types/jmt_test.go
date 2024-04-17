package types

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
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

	blob := n1.Encode()
	assert.NotNil(t, blob)
	node, err := UnmarshalJMTNodeFromPb(blob)
	assert.Nil(t, err)
	assert.True(t, equalInternalNode(n1, node.(*InternalNode)))

	n1.Children[10] = nil
	blob2 := n1.Encode()
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

	blob := leaf1.Encode()

	node, err := UnmarshalJMTNodeFromPb(blob)
	assert.Nil(t, err)

	assert.True(t, equalLeafNode(leaf1, node.(*LeafNode)))
}

func TestStateDelta_Marshal(t *testing.T) {
	stateDelta := &StateDelta{Journal: make([]*TrieJournal, 0)}
	stateDelta.Journal = append(stateDelta.Journal,
		&TrieJournal{
			Type:        1,
			RootHash:    common.HexToHash("0x4d5e85518fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef26"),
			RootNodeKey: &NodeKey{Version: 1, Path: []byte("path1"), Type: []byte("type1")},
			PruneSet: map[string]struct{}{
				"p1": {},
				"p2": {},
			},
			DirtySet: map[string]Node{
				"d1": mockRandomLeafNode(),
				"d2": mockRandomInternalNode(),
			},
		},
		&TrieJournal{
			Type:        2,
			RootHash:    common.HexToHash("0x4d5e85518fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef27"),
			RootNodeKey: &NodeKey{Version: 2, Path: []byte("path2"), Type: []byte("type2")},
			PruneSet:    map[string]struct{}{},
			DirtySet: map[string]Node{
				"d3": mockRandomLeafNode(),
			},
		},
		&TrieJournal{
			Type:        2,
			RootHash:    common.HexToHash("0x4d5e85518fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef29"),
			RootNodeKey: &NodeKey{Version: 3, Path: []byte("path3"), Type: []byte("type3")},
			PruneSet: map[string]struct{}{
				"p3": {},
			},
			DirtySet: map[string]Node{},
		},
		&TrieJournal{
			Type:        2,
			RootHash:    common.HexToHash("0x4d5e85518fb3fe5ed1eb123d4feb2a8f96b025fca63a19f02b8727d3d4f8ef28"),
			RootNodeKey: &NodeKey{Version: 4, Path: []byte("path4"), Type: []byte("type4")},
			PruneSet:    map[string]struct{}{},
			DirtySet:    map[string]Node{},
		},
	)

	blob := stateDelta.Encode()
	assert.NotNil(t, blob)
	res, err := DecodeStateDelta(blob)
	assert.Nil(t, err)
	for i := range res.Journal {
		assert.Equal(t, res.Journal[i].Type, stateDelta.Journal[i].Type)
		assert.Equal(t, res.Journal[i].RootHash, stateDelta.Journal[i].RootHash)
		assert.True(t, slices.Equal(res.Journal[i].RootNodeKey.Encode(), stateDelta.Journal[i].RootNodeKey.Encode()))
		for k := range stateDelta.Journal[i].PruneSet {
			_, ok := res.Journal[i].PruneSet[k]
			assert.True(t, ok)
		}
		for k, v1 := range stateDelta.Journal[i].DirtySet {
			v2, ok := res.Journal[i].DirtySet[k]
			assert.True(t, ok)
			assert.NotNil(t, v2)
			if leaf2, ok := v2.(*LeafNode); ok {
				leaf1, ok := v1.(*LeafNode)
				assert.True(t, ok)
				assert.True(t, equalLeafNode(leaf1, leaf2))
			} else if internal2, ok := v2.(*InternalNode); ok {
				internal1, ok := v1.(*InternalNode)
				assert.True(t, ok)
				assert.True(t, equalInternalNode(internal1, internal2))
			}
		}
	}
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

func mockRandomLeafNode() *LeafNode {
	k, v := getRandomHexKV(4, 8)
	h := sha256.Sum256(v)
	return &LeafNode{
		Key:  k,
		Val:  v,
		Hash: h,
	}
}

func mockRandomInternalNode() *InternalNode {
	_, v := getRandomHexKV(4, 8)
	rand.Seed(uint64(time.Now().UnixNano()))

	res := &InternalNode{
		Children: [16]*Child{},
	}

	res.Children[rand.Uint32()%16] = &Child{
		Version: rand.Uint64(),
		Leaf:    rand.Uint32()%2 == 1,
		Hash:    sha256.Sum256(v),
	}

	res.Children[rand.Uint32()%16] = &Child{
		Version: rand.Uint64(),
		Leaf:    rand.Uint32()%2 == 1,
		Hash:    sha256.Sum256(v),
	}

	return res
}

func getRandomHexKV(lk, lv int) (k []byte, v []byte) {
	rand.Seed(uint64(time.Now().UnixNano()))
	k = make([]byte, lk)
	v = make([]byte, lv)
	for i := 0; i < lk; i++ {
		k[i] = byte(rand.Intn(16))
	}
	for i := 0; i < lv; i++ {
		v[i] = byte(rand.Intn(16))
	}
	return k, v
}
