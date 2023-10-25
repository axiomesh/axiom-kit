package jmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
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
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	hash0 := jmt.Commit()

	jmt, err = New(hash0, s)
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
	hash1 := jmt.Commit()
	jmt, err = New(hash1, s)
	require.Nil(t, err)

	// key exist only in v1
	proof1, err := jmt.Prove(toHex("bb17"))
	require.Nil(t, err)
	exist, err = VerifyProof(hash1, proof1)
	require.Nil(t, err)
	require.True(t, exist)
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
	err := jmt.Update(0, toHex("aa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa3"), []byte("v3"))
	require.Nil(t, err)
	hash0 := jmt.Commit()

	jmt, err = New(hash0, s)
	require.Nil(t, err)

	// exist key
	proof1, err := jmt.Prove(toHex("aa1"))
	require.Nil(t, err)
	fmt.Println("proof:", proof1.deserialize())
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
	proof, err := jmt.Prove([]byte("abf3"))
	require.Nil(t, err)
	//fmt.Println("proof:", proof.print())
	hash := jmt.Commit()
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

func Test_Case_Proof_Random_1(t *testing.T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	maxn := 1000  // max number of nodes
	cases := 10   // number of test cases
	version := 10 // number of states
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
				jmt, err = New(rootHash, s)
				require.Nil(t, err)
			}
			//nnum := rand.Intn(maxn)
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
			rootHash = jmt.Commit()
			v2hash[ver] = rootHash
			jmt = nil
		}
		// verify
		for ver := 0; ver < version; ver++ {
			jmt, err = New(v2hash[ver], s)
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
