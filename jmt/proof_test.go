package jmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
)

//		                      [0_]                                              [0_]
//	                      /         \								        /           \
//					   [0_0]          [0_b]                              <0_0003>      [0_b]
//			             |              |                                                |
//	                  [0_00]          [0_bb]							               [0_bb]
//	                    |                |                    =>                         |
//	               [0_000]            ——————								           ——————
//	                   ｜           |         |									     |         |
//		   	       ——————————     <0_bb17>   <0_bbf7>	  				         <0_bb17>   <0_bbf7>
//		           |       |
//			    <0_0001>  <0_0003>
func Test_LegalProof(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte{0, 1, 2, 3, 30, 50, 80, 128, 166, 179, 200, 245, 255})
	require.Nil(t, err)
	hash0 := jmt.Commit(nil)

	jmt, err = New(hash0, s, nil, nil, jmt.logger)
	require.Nil(t, err)

	// exist key
	proof, err := jmt.Prove(toHex("0001"))
	require.Nil(t, err)
	exist, err := VerifyProof(hash0, proof)
	require.Nil(t, err)
	require.True(t, exist)
	proof, err = jmt.Prove(toHex("bb17"))
	require.Nil(t, err)
	exist, err = VerifyProof(hash0, proof)
	require.Nil(t, err)
	require.True(t, exist)

	// state transit from v0 to v1
	err = jmt.Update(1, toHex("0001"), []byte{})
	require.Nil(t, err)
	hash1 := jmt.Commit(nil)
	jmt, err = New(hash1, s, nil, nil, jmt.logger)
	require.Nil(t, err)

	// key exist only in v1
	proof, err = jmt.Prove(toHex("0001"))
	require.Nil(t, proof)
	require.Equal(t, err, ErrorInvalidPath)

	// key still exist in v2
	proof, err = jmt.Prove(toHex("bb17"))
	require.Nil(t, err)
	exist, err = VerifyProof(hash1, proof)
	require.Nil(t, err)
	require.True(t, exist)
	fmt.Printf("proof=%v\n", proof)
}

//	                   [0_]
//	                    |
//					   [0_a]
//						|
//					   [0_aa]
//						|
//	   		    ——————————————————————
//	             |        |         |
//			 <0_aa1>   <0_aa2>   <0_aa3>
func Test_IllegalProof(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("aaa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aaa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aaa3"), []byte("v3"))
	require.Nil(t, err)
	hash0 := jmt.Commit(nil)

	jmt, err = New(hash0, s, nil, nil, jmt.logger)
	require.Nil(t, err)

	// exist key
	proof1, err := jmt.Prove(toHex("aaa1"))
	require.Nil(t, err)
	exist, err := VerifyProof(hash0, proof1)
	require.Nil(t, err)
	require.True(t, exist)
	// tamper Value in proof
	copy(proof1.Value, "v2")
	exist, err = VerifyProof(hash0, proof1)
	require.Nil(t, err)
	require.False(t, exist)
	// node sequence error in merkle path
	tmp := proof1.Proof[0]
	proof1.Proof[0] = proof1.Proof[1]
	proof1.Proof[1] = tmp
	exist, err = VerifyProof(hash0, proof1)
	require.Nil(t, err)
	require.False(t, exist)
	tmp = proof1.Proof[0]
	proof1.Proof[0] = proof1.Proof[1]
	proof1.Proof[1] = tmp
	// merkle path is shorter than addressing path of Key
	proof1.Proof = proof1.Proof[:len(proof1.Proof)-1]
	exist, err = VerifyProof(hash0, proof1)
	require.Nil(t, err)
	require.False(t, exist)
}

func Test_SingleLeafNodeProof(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("abf3"), []byte("v1"))
	require.Nil(t, err)
	proof, err := jmt.Prove(toHex("abf3"))
	require.Nil(t, err)
	hash := jmt.Commit(nil)
	exist, err := VerifyProof(hash, proof)
	require.Nil(t, err)
	require.True(t, exist)
	// proof struct is illegal
	proof.Key = []byte{}
	exist, err = VerifyProof(hash, proof)
	require.NotNil(t, err)
	require.Equal(t, err, ErrorBadProof)
	require.False(t, exist)
	exist, err = VerifyProof(hash, nil)
	require.NotNil(t, err)
	require.Equal(t, err, ErrorBadProof)
	require.False(t, exist)
}

func Test_EmptyTreeProof(t *testing.T) {
	jmt, _ := initEmptyJMT()
	proof, err := jmt.Prove([]byte("key"))
	require.Nil(t, err)
	require.Nil(t, proof.Key)
	require.Nil(t, proof.Value)
	require.Nil(t, proof.Proof)
}

func Test_VerifyEmptyTrie(t *testing.T) {
	jmt, backend := initEmptyJMT()
	rootHash := jmt.Commit(nil)
	verified, err := VerifyTrie(rootHash, backend, nil)
	require.Equal(t, ErrorNotFound, err)
	require.False(t, verified)
}

//		                           [0_]
//	                           /         \
//						    [0_0]          [0_b]
//							  |              |
//	                       [0_00]          [0_bb]
//	                         |                |
//	                    [0_000]            ——————
//	                        ｜           |         |
//		   	            ——————————     <0_bb17>   <0_bbf7>
//		                |       |
//			         <0_0001>  <0_0003>
func Test_VerifyIllegalTrie(t *testing.T) {
	jmt, backend := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)

	nk1 := &types.NodeKey{
		Version: 0,
		Path:    toHex("bb"),
		Type:    jmt.typ,
	}
	_, ok := jmt.dirtySet[string(nk1.Encode())].(*types.InternalNode)
	require.True(t, ok)

	nk2 := &types.NodeKey{
		Version: 0,
		Path:    toHex("bbf"),
		Type:    jmt.typ,
	}
	_, ok = jmt.dirtySet[string(nk2.Encode())].(*types.LeafNode)
	require.True(t, ok)

	nk3 := &types.NodeKey{
		Version: 0,
		Path:    toHex("000"),
		Type:    jmt.typ,
	}
	_, ok = jmt.dirtySet[string(nk3.Encode())].(*types.InternalNode)
	require.True(t, ok)

	nk4 := &types.NodeKey{
		Version: 0,
		Path:    toHex("b"),
		Type:    jmt.typ,
	}
	_, ok = jmt.dirtySet[string(nk4.Encode())].(*types.InternalNode)
	require.True(t, ok)

	rootHash := jmt.Commit(nil)

	t.Run("test internal node content invalid", func(t *testing.T) {
		n1, _ := jmt.getNode(nk1)
		n1.(*types.InternalNode).Children[1].Version = 1 // modify InternalNode's content
		backend.Put(nk1.Encode(), n1.Encode())

		verified, err := VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.False(t, verified)

		n11 := n1.Copy().(*types.InternalNode)
		n11.Children[1].Version = 0
		backend.Put(nk1.Encode(), n11.Encode())
		verified, err = VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.True(t, verified)

		backend.Delete(nk1.Encode())
		verified, err = VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.False(t, verified)
		backend.Put(nk1.Encode(), n11.Encode())
	})

	t.Run("test leaf node content invalid", func(t *testing.T) {
		n2, _ := jmt.getNode(nk2)
		originVal := n2.(*types.LeafNode).Val
		n2.(*types.LeafNode).Val = []byte{1} // modify LeafNode's content
		backend.Put(nk2.Encode(), n2.Encode())
		verified, err := VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.False(t, verified)

		n2.(*types.LeafNode).Val = originVal
		backend.Put(nk2.Encode(), n2.Encode())

		verified, err = VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.True(t, verified)
	})

	t.Run("test leaf node hash invalid", func(t *testing.T) {
		n2, _ := jmt.getNode(nk2)
		originHash := n2.(*types.LeafNode).Hash
		n2.(*types.LeafNode).Hash = common.BytesToHash([]byte{1}) // modify LeafNode's content
		backend.Put(nk2.Encode(), n2.Encode())
		verified, err := VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.False(t, verified)

		n2.(*types.LeafNode).Hash = originHash
		backend.Put(nk2.Encode(), n2.Encode())

		verified, err = VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.True(t, verified)
	})

	t.Run("test internal node of different branch content invalid", func(t *testing.T) {
		n3, _ := jmt.getNode(nk3)
		n3.(*types.InternalNode).Children[0] = n3.(*types.InternalNode).Children[1] // modify InternalNode's content
		backend.Put(nk3.Encode(), n3.Encode())

		n4, _ := jmt.getNode(nk4)
		n4.(*types.InternalNode).Children[11].Version = 1 // modify InternalNode's content
		backend.Put(nk4.Encode(), n4.Encode())

		verified, err := VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.False(t, verified)

		n33 := n3.Copy().(*types.InternalNode)
		n33.Children[0] = nil
		backend.Put(nk3.Encode(), n33.Encode())
		n44 := n4.Copy().(*types.InternalNode)
		n44.Children[11].Version = 0
		backend.Put(nk4.Encode(), n44.Encode())
		verified, err = VerifyTrie(rootHash, backend, nil)
		require.Nil(t, err)
		require.True(t, verified)
	})
}

func Test_Case_Proof_Random_1(t *testing.T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	maxn := 1000  // max number of nodes
	cases := 10   // number of test cases
	version := 10 // number of states
	logger := log.NewWithModule("JMT-Test")
	for i := 0; i < cases; i++ {
		fmt.Println("【Random Testcase】 ", i)
		jmt, s := initEmptyJMT()
		rootHash := placeHolder
		var err error
		v2inserted := make(map[int]map[string][]byte, version)
		v2deleted := make(map[int]map[string]struct{}, version)
		v2hash := make(map[int]common.Hash, version)
		for ver := 0; ver < version; ver++ {
			if jmt == nil {
				jmt, err = New(rootHash, s, nil, nil, logger)
				require.Nil(t, err)
			}
			// nnum := rand.Intn(maxn)
			nnum := maxn
			inserted := make(map[string][]byte, nnum)
			deleted := make(map[string]struct{}, nnum)
			// random insert
			for j := 0; j < nnum; j++ {
				k, v := getRandomHexKV(4, 16)
				inserted[string(k)] = v
				err = jmt.Update(uint64(ver), k, v)
				require.Nil(t, err)
			}
			// random delete
			for j := 0; j < nnum/2; j++ {
				k, _ := getRandomHexKV(4, 16)
				deleted[string(k)] = struct{}{}
				delete(inserted, string(k))
				err = jmt.Update(uint64(ver), k, []byte{})
				require.Nil(t, err)
			}
			v2inserted[ver] = inserted
			v2deleted[ver] = deleted
			rootHash = jmt.Commit(nil)
			v2hash[ver] = rootHash
			jmt = nil
		}
		// verify
		for ver := 0; ver < version; ver++ {
			jmt, err = New(v2hash[ver], s, nil, nil, logger)
			require.Nil(t, err)
			for k, v := range v2inserted[ver] {
				proof, err := jmt.Prove([]byte(k))
				require.Nil(t, err)
				exist, err := VerifyProof(v2hash[ver], proof)
				require.True(t, slices.Equal(proof.Value, v))
				require.Nil(t, err)
				require.True(t, exist)
			}
		}
	}
}
