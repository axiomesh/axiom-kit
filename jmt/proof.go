package jmt

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

var (
	ErrorBadProof    = errors.New("proof is invalid")    // proof content or struct is illegal
	ErrorInvalidPath = errors.New("invalid merkle path") // addressing merkle path error whe generate proof
	ErrorNodeMissing = errors.New("node is missing")     // miss node in merkle path
)

type ProofResult struct {
	Key   []byte
	Value []byte
	Proof [][]byte // merkle path from top to bottom, the first element is root
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
		proof.Proof = append(proof.Proof, n.Encode())
		if n.Children[key[next]] == nil {
			return ErrorInvalidPath
		}
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey := &types.NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Type:    jmt.typ,
		}
		var nextNode types.Node
		nextNode, err := jmt.getNode(nextNodeKey)
		if err != nil {
			return err
		}
		return jmt.prove(nextNode, key, next+1, proof) // find in next layer in tree
	case *types.LeafNode:
		if !bytes.Equal(n.Key, key) {
			return ErrorInvalidPath
		}
		proof.Proof = append(proof.Proof, n.Encode())
		proof.Key = n.Key
		proof.Value = n.Val
		return nil
	default:
		return nil
	}
}

// VerifyTrie verifies a whole trie
func VerifyTrie(rootHash common.Hash, backend storage.Storage, cache PruneCache) (bool, error) {
	logger := log.NewWithModule("JMT-VerifyTrie")

	trie, err := New(rootHash, backend, nil, cache, logger)
	if err != nil {
		return false, err
	}
	if trie.root == nil {
		return false, ErrorNotFound
	}

	switch n := (trie.root).(type) {
	case *types.InternalNode:
		var i, cnt byte
		resChan := make(chan bool, 16)
		defer close(resChan)

		wg := sync.WaitGroup{}
		for i, cnt = 0, 0; i < 16; i++ {
			child := n.Children[i]
			if n.Children[i] == nil {
				continue
			}

			nextPath := []byte{i}
			nextNodeKey := &types.NodeKey{
				Version: child.Version,
				Path:    nextPath,
				Type:    trie.typ,
			}
			nextNode, err := trie.getNode(nextNodeKey)
			if err != nil {
				return false, err
			}
			wg.Add(1)
			cnt++
			go func() {
				defer wg.Done()
				verified, err := trie.verifySubTrie(nextNode, child.Hash, nextPath)
				if err != nil {
					logger.Errorf("verifySubTrie error: %v", err.Error())
				}
				resChan <- verified
			}()
		}
		wg.Wait()

		for i = 0; i < cnt; i++ {
			verified, ok := <-resChan
			if ok && !verified {
				return false, nil
			}
		}
		return true, nil
	case *types.LeafNode:
		if n.Hash != rootHash {
			return false, nil
		}
		return true, nil
	default:
		panic("unsupported trie node type")
	}
}

// verifySubTrie verifies a sub-trie recursively
func (jmt *JMT) verifySubTrie(root types.Node, rootHash common.Hash, path []byte) (bool, error) {
	if root == nil {
		return false, ErrorNodeMissing
	}
	if root.GetHash() != rootHash {
		jmt.logger.Errorf("[verifySubTrie] target node: %v, expected hash: %v", root, rootHash)
		return false, nil
	}
	switch n := (root).(type) {
	case *types.InternalNode:
		var i byte
		for i = 0; i < 16; i++ {
			if n.Children[i] == nil {
				continue
			}
			nextPath := make([]byte, len(path))
			copy(nextPath, path)
			nextPath = append(nextPath, i)
			nextNodeKey := &types.NodeKey{
				Version: n.Children[i].Version,
				Path:    nextPath,
				Type:    jmt.typ,
			}
			var nextNode types.Node
			nextNode, err := jmt.getNode(nextNodeKey)
			if err != nil {
				return false, err
			}
			verified, err := jmt.verifySubTrie(nextNode, n.Children[i].Hash, nextPath)
			if !verified || err != nil {
				return false, err
			}
		}
		return true, nil
	case *types.LeafNode:
		if n.Hash != rootHash {
			jmt.logger.Errorf("[verifySubTrie] target leaf node: %v, expected hash: %v", root, rootHash)
			return false, nil
		}
		return true, nil
	default:
		panic("unsupported trie node type")
	}
}

// VerifyProof support key existence proof
func VerifyProof(rootHash common.Hash, proof *ProofResult) (bool, error) {
	if proof == nil || len(proof.Key) == 0 {
		return false, ErrorBadProof
	}
	return verifyProof(rootHash, 0, proof)
}

func verifyProof(hash common.Hash, level int, proof *ProofResult) (bool, error) {
	if level >= len(proof.Proof) {
		return false, nil
	}
	n, err := types.UnmarshalJMTNodeFromPb(proof.Proof[level])
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
		return verifyProof(nn.Children[proof.Key[level]].Hash, level+1, proof) // verify node in next layer
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
func (proof *ProofResult) String() string {
	res := strings.Builder{}
	res.WriteString("\nKey:[")
	res.WriteString(base64.StdEncoding.EncodeToString(proof.Key))
	res.WriteString("],\nValue:[")
	res.WriteString(base64.StdEncoding.EncodeToString(proof.Value))
	res.WriteString("],\nMerkle Path:[")

	for _, s := range proof.Proof {
		res.WriteString("\"")
		res.WriteString(base64.StdEncoding.EncodeToString(s))
		res.WriteString("\",\n")
	}
	res.WriteString("]")
	return res.String()
}
