package jmt

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
)

type TraversedNode struct {
	Origin *types.Node
	Path   []byte
}

type (
	NodeKey struct {
		Version uint64 // version of current tree node.
		Path    []byte // addressing path from root to current node. Path is part of LeafNode.Key.
		Type    []byte // additional field for identify a tree uniquely together with Version and Path.
	}

	NodeKeyHeap []*NodeKey // max heap to store NodeKey
)

// just for debug
func (nk NodeKey) print() string {
	res := strings.Builder{}
	res.WriteString("Version[")
	res.WriteString(strconv.Itoa(int(nk.Version)))
	res.WriteString("], Path[")
	res.WriteString(hexutil.DecodeFromNibbles(nk.Path))
	res.WriteString("]")
	return res.String()
}

func (nk *NodeKey) Encode() []byte {
	return nk.encode()
}

// encode NodeKey to bytes for physical storage
func (nk *NodeKey) encode() []byte {
	var length byte
	for i := 0; i < len(nk.Type); i++ {
		length++
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, nk.Version)
	buf = append(buf, length)
	buf = append(buf, nk.Type...)
	buf = append(buf, nk.Path...)
	return buf
}

// decodeNodeKey print from bytes in physical storage to NodeKey
func decodeNodeKey(raw []byte) *NodeKey {
	nk := &NodeKey{}
	nk.Version = binary.BigEndian.Uint64(raw[:8])
	length := raw[8]
	nk.Type = raw[9 : 9+length]
	nk.Path = raw[9+length:]
	return nk
}

func (h NodeKeyHeap) Len() int {
	return len(h)
}

func (h NodeKeyHeap) Less(i, j int) bool {
	if h[i].Version != h[j].Version {
		return h[i].Version > h[j].Version
	}

	if !bytes.Equal(h[i].Type, h[j].Type) {
		return bytes.Compare(h[i].Type, h[j].Type) > 0
	}

	return bytes.Compare(h[i].Path, h[j].Path) < 0
}

func (h NodeKeyHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *NodeKeyHeap) Push(x interface{}) {
	*h = append(*h, x.(*NodeKey))
}

func (h *NodeKeyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
