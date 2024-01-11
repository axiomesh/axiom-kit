package types

import (
	"crypto/sha256"
	"strconv"
	"strings"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types/pb"
	"github.com/ethereum/go-ethereum/common"
)

type Node interface {
	Encode() []byte
	GetHash() common.Hash
	Copy() Node
	Print() string // just for debug
}

type (
	InternalNode struct {
		Children [16]*Child
	}

	LeafNode struct {
		Key  []byte
		Val  []byte
		Hash common.Hash
	}

	Child struct {
		Hash    common.Hash
		Version uint64
		Leaf    bool
	}
)

func (n *InternalNode) Encode() []byte {
	blob, err := n.marshal()
	if err != nil {
		return nil
	}
	return blob
}

func (n *LeafNode) Encode() []byte {
	blob, err := n.marshal()
	if err != nil {
		return nil
	}
	return blob
}

// just for debug
func (n *InternalNode) Print() string {
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
func (n *LeafNode) Print() string {
	res := strings.Builder{}
	res.WriteString("Key[")
	res.WriteString(hexutil.DecodeFromNibbles(n.Key))
	res.WriteString("], ValueLen[")
	res.WriteString(strconv.Itoa(len(n.Val)))
	res.WriteString("], Hash[")
	res.WriteString(n.Hash.String()[2:6])
	res.WriteString("]")
	return res.String()
}

func (n *InternalNode) GetHash() common.Hash {
	d, _ := n.marshal()
	data := sha256.Sum256(d)
	return data
}

// GetHash get LeafNode's hash, which is only determined by its Key and Val
func (n *LeafNode) GetHash() common.Hash {
	tmp := &LeafNode{
		Hash: common.Hash{},
		Key:  n.Key,
		Val:  n.Val,
	}
	d, _ := tmp.marshal()
	data := sha256.Sum256(d)
	return data
}

func (n *InternalNode) Copy() Node {
	return &InternalNode{
		Children: n.Children,
	}
}

func (n *LeafNode) Copy() Node {
	nn := &LeafNode{
		Hash: n.Hash,
	}
	copy(nn.Key, n.Key)
	copy(nn.Val, n.Val)
	return nn
}

func (n *InternalNode) marshal() ([]byte, error) {
	if n == nil {
		return nil, nil
	}

	children := make([]*pb.Child, 16)
	for i, child := range n.Children {
		if child == nil {
			continue
		}
		children[i] = &pb.Child{
			Version: child.Version,
			Leaf:    child.Leaf,
			Hash:    child.Hash[:],
		}
	}

	blob := &pb.InternalNode{
		Children: children,
	}

	content, err := blob.MarshalVTStrict()
	if err != nil {
		return nil, err
	}

	node := &pb.Node{
		Leaf:    false,
		Content: content,
	}
	return node.MarshalVTStrict()
}

func (n *InternalNode) fromPB(p *pb.InternalNode) {
	n.Children = [16]*Child{}
	for i, child := range p.Children {
		if len(child.Hash) == 0 {
			continue
		}
		n.Children[i] = &Child{
			Hash:    common.BytesToHash(child.Hash),
			Version: child.Version,
			Leaf:    child.Leaf,
		}
	}
}

func (n *InternalNode) unmarshal(data []byte) error {
	helper := pb.InternalNodeFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}
	n.fromPB(helper)
	return nil
}

func (n *LeafNode) marshal() ([]byte, error) {
	if n == nil {
		return nil, nil
	}

	blob := &pb.LeafNode{
		Key:   hexToBytes(n.Key),
		Value: n.Val,
		Hash:  n.Hash[:],
	}

	content, err := blob.MarshalVTStrict()
	if err != nil {
		return nil, err
	}

	node := &pb.Node{
		Leaf:    true,
		Content: content,
	}
	return node.MarshalVTStrict()
}

func (n *LeafNode) fromPB(p *pb.LeafNode) {
	n.Val = p.Value
	n.Key = bytesToHex(p.Key)
	n.Hash = common.BytesToHash(p.Hash)
}

func (n *LeafNode) unmarshal(data []byte) error {
	helper := pb.LeafNodeFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}
	n.fromPB(helper)
	return nil
}

func UnmarshalJMTNode(data []byte) (Node, error) {
	if len(data) == 0 {
		return nil, nil
	}
	helper := pb.NodeFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return nil, err
	}

	if helper.Leaf {
		res := &LeafNode{}
		err = res.unmarshal(helper.Content)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	if !helper.Leaf {
		res := &InternalNode{}
		err = res.unmarshal(helper.Content)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	return nil, nil
}

// bytesToHex expand normal bytes to hex bytes (nibbles)
func bytesToHex(h []byte) []byte {
	if len(h) == 0 {
		return h
	}
	var i byte = 0
	length := h[0]
	tmp := h[1:]
	dst := make([]byte, length)
	for ; i < length; i++ {
		if i&0x01 == 0 {
			dst[i] = tmp[i/2] >> 4
		} else {
			dst[i] = tmp[i/2] & 15
		}
	}
	return dst
}

// hexToBytes compress hex bytes (also called nibbles) to normal bytes
func hexToBytes(src []byte) []byte {
	if len(src) == 0 {
		return src
	}
	if len(src) > 255 {
		panic("don't support compress bytes with length greater than 255 ")
	}
	length := (1 + len(src)) / 2
	res := make([]byte, 1+length)
	res[0] = byte(len(src))
	for i := 0; i < length; i++ {
		if 2*i+1 == len(src) {
			res[i+1] = src[2*i] << 4
		} else {
			res[i+1] = (src[2*i] << 4) | src[2*i+1]
		}
	}
	return res
}
