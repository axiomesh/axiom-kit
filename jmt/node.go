package jmt

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/hexutil"
)

type Node interface {
	encode() []byte // todo use customized encoding method (RLP)
	hash() common.Hash
	copy() Node
	Print() string // just for debug
}

type TraversedNode struct {
	Origin *Node
	Path   []byte
}

type (
	InternalNode struct {
		Children [16]*Child `json:"children"`
	}

	LeafNode struct {
		Key  []byte      `json:"key"`
		Val  []byte      `json:"value"`
		Hash common.Hash `json:"hash"`
	}

	Child struct {
		Hash    common.Hash `json:"hash"`
		Version uint64      `json:"version"`
		Leaf    bool        `json:"leaf"`
	}

	NodeKey struct {
		Version uint64 // version of current tree node.
		Path    []byte // addressing path from root to current node. Path is part of LeafNode.Key.
		Prefix  []byte // additional field for identify a tree uniquely together with Version and Path.
	}
)

func (n InternalNode) encode() []byte {
	b, _ := json.Marshal(n)
	return b
}

func (n LeafNode) encode() []byte {
	b, _ := json.Marshal(n)
	return b
}

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

// just for debug
func (n InternalNode) Print() string {
	res := strings.Builder{}
	for i := 0; i < 16; i++ {
		res.WriteString(strconv.Itoa(i))
		if n.Children[i] == nil {
			res.WriteString(", ")
		} else {
			child := n.Children[i]
			res.WriteString(":<Hash[")
			res.WriteString(child.Hash.String()[2:6])
			res.WriteString("], Version[")
			res.WriteString(strconv.Itoa(int(child.Version)))
			res.WriteString("], Leaf[")
			res.WriteString(strconv.FormatBool(child.Leaf))
			res.WriteString("]>, ")
		}
	}
	return res.String()
}

// just for debug
func (n LeafNode) Print() string {
	res := strings.Builder{}
	res.WriteString("Key[")
	res.WriteString(hexutil.DecodeFromNibbles(n.Key))
	res.WriteString("], Value[")
	res.Write(n.Val)
	res.WriteString("], Hash[")
	res.WriteString(n.Hash.String()[2:6])
	res.WriteString("]")
	return res.String()
}

func (n InternalNode) hash() common.Hash {
	d := n.encode()
	data := sha256.Sum256(d)
	return data
}

// hash of LeafNode is only determined by its Key and Val
func (n LeafNode) hash() common.Hash {
	n.Hash = common.Hash{}
	d := n.encode()
	data := sha256.Sum256(d)
	return data
}

// deep copy
func (n InternalNode) copy() Node {
	return InternalNode{
		Children: n.Children,
	}
}

// deep copy
func (n LeafNode) copy() Node {
	nn := LeafNode{
		Hash: n.Hash,
	}
	copy(nn.Key, n.Key)
	copy(nn.Val, n.Val)
	return nn
}

func decodeNode(rawNode []byte) (n Node, err error) {
	if len(rawNode) == 0 {
		return nil, nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(rawNode, &m); err != nil {
		return nil, err
	}
	if m["children"] != nil {
		return decodeInternalNode(rawNode)
	} else if m["key"] != nil {
		return decodeLeafNode(rawNode)
	}
	return nil, nil
}

func decodeInternalNode(rawNode []byte) (Node, error) {
	n := InternalNode{}
	err := json.Unmarshal(rawNode, &n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func decodeLeafNode(rawNode []byte) (Node, error) {
	n := LeafNode{}
	err := json.Unmarshal(rawNode, &n)
	if err != nil {
		return nil, err
	}
	return n, nil
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
