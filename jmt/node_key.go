package jmt

import (
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
		Prefix  []byte // additional field for identify a tree uniquely together with Version and Path.
	}
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
	for i := 0; i < len(nk.Prefix); i++ {
		length++
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, nk.Version)
	buf = append(buf, length)
	buf = append(buf, nk.Prefix...)
	buf = append(buf, nk.Path...)
	return buf
}

// decodeNodeKey print from bytes in physical storage to NodeKey
func decodeNodeKey(raw []byte) *NodeKey {
	nk := &NodeKey{}
	nk.Version = binary.BigEndian.Uint64(raw[:8])
	length := raw[8]
	nk.Prefix = raw[9 : 9+length]
	nk.Path = raw[9+length:]
	return nk
}
