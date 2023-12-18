package jmt

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types"
)

var (
	ErrorBadProof = errors.New("proof is invalid") // proof content or struct is illegal
)

type ProofResult struct {
	Key   []byte   `json:"key"`
	Value []byte   `json:"value"`
	Proof []string `json:"proof"` // merkle path from top to bottom, the first element is root
}

func (jmt *JMT) Prove(key []byte) (*ProofResult, error) {
	proof := &ProofResult{}
	err := jmt.prove(jmt.root, key, 0, proof)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

func (jmt *JMT) prove(root types.Node, key []byte, next int, proof *ProofResult) error {
	switch n := (root).(type) {
	case *types.InternalNode:
		proof.Proof = append(proof.Proof, string(n.Encode()))
		if n.Children[key[next]] == nil {
			return nil
		}
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey := &NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Prefix:  jmt.prefix,
		}
		var nextNode types.Node
		nextNode, err := jmt.getNode(nextNodeKey)
		if err != nil {
			return err
		}
		return jmt.prove(nextNode, key, next+1, proof) // find in next layer in tree
	case *types.LeafNode:
		proof.Proof = append(proof.Proof, string(n.Encode()))
		proof.Key = n.Key
		proof.Value = n.Val
		return nil
	default:
		return nil
	}
}

// VerifyProof support key existence proof
func VerifyProof(rootHash common.Hash, proof *ProofResult) (bool, error) {
	if proof == nil || len(proof.Key) == 0 {
		return false, ErrorBadProof
	}
	return verify(rootHash, 0, proof)
}

func verify(hash common.Hash, level int, proof *ProofResult) (bool, error) {
	if level >= len(proof.Proof) {
		return false, nil
	}
	n, err := types.UnmarshalJMTNode([]byte(proof.Proof[level]))
	if err != nil {
		return false, ErrorBadProof
	}
	switch nn := (n).(type) {
	case *types.InternalNode:
		if nn.Children[proof.Key[level]] == nil {
			return false, nil
		}
		if hash != nn.GetHash() { // verify current node's hash is include in proof
			return false, nil
		}
		return verify(nn.Children[proof.Key[level]].Hash, level+1, proof) // verify node in next layer
	case *types.LeafNode:
		leaf := types.LeafNode{
			Key: proof.Key,
			Val: proof.Value,
		}
		if leaf.GetHash() != hash { // verify whether current node's hash were included in proof
			return false, nil
		}
		return true, nil
	default:
		return false, ErrorBadProof
	}
}

// just for debug
func (proof *ProofResult) deserialize() string {
	res := strings.Builder{}
	res.WriteString("\nKey:[")
	res.WriteString(hexutil.DecodeFromNibbles(proof.Key))
	res.WriteString("],\nValue:[")
	res.WriteString(string(proof.Value))
	res.WriteString("],\nMerkle Path:[")

	for _, s := range proof.Proof {
		res.WriteString("\"")
		res.WriteString(s)
		res.WriteString("\",\n")
	}
	res.WriteString("]")
	return res.String()
}
