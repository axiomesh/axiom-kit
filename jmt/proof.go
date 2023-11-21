package jmt

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/hexutil"
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

func (jmt *JMT) prove(root Node, key []byte, next int, proof *ProofResult) error {
	switch n := (root).(type) {
	case InternalNode:
		proof.Proof = append(proof.Proof, string(n.encode()))
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
		var nextNode Node
		nextNode, err := jmt.getNode(nextNodeKey)
		if err != nil {
			return err
		}
		return jmt.prove(nextNode, key, next+1, proof) // find in next layer in tree
	case LeafNode:
		proof.Proof = append(proof.Proof, string(n.encode()))
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
	n, err := decodeNode([]byte(proof.Proof[level]))
	if err != nil {
		return false, ErrorBadProof
	}
	switch nn := (n).(type) {
	case InternalNode:
		if nn.Children[proof.Key[level]] == nil {
			return false, nil
		}
		if hash != nn.hash() { // verify current node's hash is include in proof
			return false, nil
		}
		return verify(nn.Children[proof.Key[level]].Hash, level+1, proof) // verify node in next layer
	case LeafNode:
		leaf := LeafNode{
			Key: proof.Key,
			Val: proof.Value,
		}
		if leaf.hash() != hash { // 验证本节点的哈希包含在proof内
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
