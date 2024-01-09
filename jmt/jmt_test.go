package jmt

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/pebble"
)

func Test_EmptyTree(t *testing.T) {
	jmt, _ := initEmptyJMT()
	n, err := jmt.Get([]byte("key"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_EmptyKey(t *testing.T) {
	jmt, _ := initEmptyJMT()
	n, err := jmt.Get(nil)
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get([]byte{})
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_SingleTreeNode(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))

	rootHash := jmt.Commit()
	require.Equal(t, rootHash, jmt.root.GetHash())

	jmt, err = New(rootHash, s)
	require.Nil(t, err)

	// get from kv
	n, err = jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
}

//	                   [0_]
//	                    |
//					   [0_a]
//						|
//	   		    ——————————————————————
//	             |        |         |
//			 <0_a1>   <0_a2>   <0_a3>
func Test_AllLeafNodeWith1SamePrefix(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a3"), []byte("v3"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("a2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("a3"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
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
func Test_AllLeafNodeWith2SamePrefix(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("aa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa3"), []byte("v3"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("aa1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("aa2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("aa3"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
}

//		                     [0_]
//		                      |
//						    [0_a]
//							  |
//	                     [0_aa]
//	                   /         \
//		   		    ——————————      <0_ab3>
//		             |       |
//			     <0_aa1>  <0_aa2>
func Test_InternalNodeHasTwoTypesOfChildNode(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("aa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("ab3"), []byte("v3"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("aa1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("aa2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("ab3"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
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
func Test_ForkFromRoot(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
}

func Test_UpdateAfterInsert(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0001"), []byte("v3"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
}

func Test_GetNonExistKey(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("abcd"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_LoadNonExistTree(t *testing.T) {
	s := initKV()
	jmt, err := New(common.Hash{}, s)
	require.Nil(t, jmt)
	require.NotNil(t, err)
	require.Equal(t, err, ErrorNotFound)
}

func Test_DeleteUntilSingleLeafNode(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("aa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("aa1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("aa2"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_DeleteUntilEmptyTree(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a1"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_DeleteUntilEmptyTreeWithStateTransit1(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a1"), nil)
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("a2"))
	require.Nil(t, err)
	require.Equal(t, n, []byte("v2"))
	rootHash := jmt.Commit()

	// state transit
	jmt1, err := New(rootHash, s)
	require.Nil(t, err)
	n, err = jmt1.Get(toHex("a2"))
	require.Nil(t, err)
	require.Equal(t, n, []byte("v2"))
	err = jmt1.Update(1, toHex("a2"), nil)
	require.Nil(t, err)
	jmt1.Commit()
	n, err = jmt1.Get(toHex("a1"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt1.Get(toHex("a2"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_DeleteUntilEmptyTreeWithStateTransit2(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a2"), []byte("v2"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("a2"))
	require.Nil(t, err)
	require.Equal(t, n, []byte("v2"))
	rootHash := jmt.Commit()

	// state transit
	jmt1, err := New(rootHash, s)
	require.Nil(t, err)
	err = jmt1.Update(1, toHex("a1"), nil)
	require.Nil(t, err)
	err = jmt1.Update(1, toHex("a2"), nil)
	require.Nil(t, err)
	rootHash1 := jmt1.Commit()

	// verify
	jmt2, err := New(rootHash1, s)
	n, err = jmt2.Get(toHex("a1"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt2.Get(toHex("a2"))
	require.Nil(t, err)
	require.Nil(t, n)
}

func Test_DeleteNonExistKey(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("a1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a2"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("a3"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("a1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("a2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("a3"))
	require.Nil(t, err)
	require.Nil(t, n)
}

//	                   [0_]                                   [0_]
//	                    |                                      |
//					   [0_a]                                 [0_a]
//						|                                      |
//					   [0_aa]                 =>             [0_aa]
//						|                                      |
//	   		    ——————————————————————                   ——————————————
//	             |        |         |                    |            |
//			 <0_aa1>   <0_aa2>   <0_aa3>               <0_aa1>       <0_aa3>
func Test_NoCompactInternalNodeAfterDelete(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("aa1"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa3"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("aa2"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("aa1"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("aa2"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("aa3"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
}

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
func Test_CompactInternalNodeAfterDelete(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0001"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
}

func Test_DeleteFromEmptyTree(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte{})
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.Nil(t, n)
	rootHash := jmt.Commit()
	jmt, err = New(rootHash, s)
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.Nil(t, n)
}

//			                      [0_]                                                  [0_]
//		                      /         \								        /                 \
//						   [0_0]          [0_b]                              <0_0>               <0_bbf7>
//				             |              |                                   |
//		                  [0_00]          [0_bb]							 [0_00]
//		                    |                |                    =>           |
//		               [0_000]            ——————							 ——————
//		                   ｜           |         |						   |	      |
//			   	       ——————————     <0_bb17>   <0_bbf7>	  		   [0_000]      <0_0035>
//			           |       |                                           |
//				    <0_0001>  <0_0003>                                 —————————
//	                                                               |         |
//	                                                             <0_0001>    <0_0003>
func Test_ReInsertAfterDelete(t *testing.T) {
	jmt, _ := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0001"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0001"), []byte("v5"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0035"), []byte("v6"))
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v5"))
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0035"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v6"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.Nil(t, n)
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
func Test_GetAfterCommit(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	rootHash := jmt.Commit()
	require.Equal(t, rootHash, jmt.root.GetHash())
	jmt, err = New(rootHash, s)
	require.Nil(t, err)
	// get from kv
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
	// get from cache
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
}

func Test_DeleteExistKeyAndCommit(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte{})
	require.Nil(t, err)
	rootHash := jmt.Commit()
	require.Equal(t, rootHash, jmt.root.GetHash())
	jmt, err = New(rootHash, s)
	require.Nil(t, err)
	// get from kv
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.Nil(t, n)
}

//			    (V0)              [0_]                               (V1)              [0_]
//		                      /         \								        /                 \
//						   [0_0]          [0_b]                              <0_0>               <0_bbf7>
//				             |              |                                   |
//		                  [0_00]          [0_bb]							 [0_00]
//		                    |                |                    =>           |
//		               [0_000]            ——————							 ——————
//		                   ｜           |         |						   |	      |
//			   	       ——————————     <0_bb17>   <0_bbf7>	  		   [0_000]      <0_0035>
//			           |       |                                           |
//				    <0_0001>  <0_0003>                                 —————————
//	                                                               |         |
//	                                                             <0_0001>    <0_0003>
func Test_StateTransit(t *testing.T) {
	// init version 0 jmt
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	// commit version 0 jmt, and load it from kv
	rootHash0 := jmt.Commit()
	require.Equal(t, rootHash0, jmt.root.GetHash())
	jmt, err = New(rootHash0, s)
	require.Nil(t, err)
	// transit from v0 to v1
	err = jmt.Update(1, toHex("0001"), []byte("v5"))
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bbf7"), []byte("v6"))
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bb17"), []byte("v8"))
	require.Nil(t, err)
	// commit version 1 jmt, and load it from kv
	rootHash1 := jmt.Commit()
	require.Equal(t, rootHash1, jmt.root.GetHash())
	jmt, err = New(rootHash1, s)
	require.Nil(t, err)
	// transit from v1 to v2
	err = jmt.Update(2, toHex("0001"), []byte("v9"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bbf7"), []byte("v10"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bb17"), []byte("v12"))
	require.Nil(t, err)
	rootHash2 := jmt.Commit()
	require.Equal(t, rootHash2, jmt.root.GetHash())
	// verify v0
	jmt, err = New(rootHash0, s)
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	require.Nil(t, err)
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
	// verify v1
	jmt, err = New(rootHash1, s)
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v5"))
	require.Nil(t, err)
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v6"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v8"))
	// verify v2
	jmt, err = New(rootHash2, s)
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v9"))
	require.Nil(t, err)
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v10"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v12"))
}

func Test_StateTransitWithDelete(t *testing.T) {
	// init version 0 jmt
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("0003"), []byte("v3"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bb17"), []byte("v4"))
	require.Nil(t, err)
	// commit version 0 jmt, and load it from kv
	rootHash0 := jmt.Commit()
	require.Equal(t, rootHash0, jmt.root.GetHash())
	jmt, err = New(rootHash0, s)
	require.Nil(t, err)
	// transit from v0 to v1
	err = jmt.Update(1, toHex("0001"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bbf7"), []byte("v5"))
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bb17"), []byte("v6"))
	require.Nil(t, err)
	// commit version 1 jmt, and load it from kv
	rootHash1 := jmt.Commit()
	require.Equal(t, rootHash1, jmt.root.GetHash())
	jmt, err = New(rootHash1, s)
	require.Nil(t, err)
	// transit from v1 to v2
	err = jmt.Update(2, toHex("0001"), []byte("v7"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bbf7"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bb17"), []byte("v8"))
	require.Nil(t, err)
	rootHash2 := jmt.Commit()
	require.Equal(t, rootHash2, jmt.root.GetHash())
	// verify v0
	jmt, err = New(rootHash0, s)
	require.Nil(t, err)
	n, err := jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v1"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
	// verify v1
	jmt, err = New(rootHash1, s)
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v5"))
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v6"))
	// verify v2
	jmt, err = New(rootHash2, s)
	require.Nil(t, err)
	n, err = jmt.Get(toHex("0001"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v7"))
	n, err = jmt.Get(toHex("bbf7"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt.Get(toHex("0003"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt.Get(toHex("bb17"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v8"))
}

func Test_KeyLengthEqualTo1(t *testing.T) {
	// init version 0 jmt
	jmt0, s0 := initEmptyJMT()
	err := jmt0.Update(0, toHex("2"), []byte("v2"))
	require.Nil(t, err)
	rootHash0 := jmt0.Commit()

	// transit from v0 to v1
	jmt1, err := New(rootHash0, s0)
	require.Nil(t, err)
	err = jmt1.Update(1, toHex("2"), []byte("v3"))
	require.Nil(t, err)
	err = jmt1.Update(1, toHex("4"), []byte("v4"))
	require.Nil(t, err)
	rootHash1 := jmt1.Commit()

	// verify v0
	jmt0, err = New(rootHash0, s0)
	require.Nil(t, err)
	n, err := jmt0.Get(toHex("2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))

	// verify v1
	jmt1, err = New(rootHash1, s0)
	require.Nil(t, err)
	n, err = jmt1.Get(toHex("2"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt1.Get(toHex("4"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))
}

func Test_StateTransitWithDifferentInsertOrder(t *testing.T) {
	// init version 0 jmt
	jmt1, s1 := initEmptyJMT()
	err := jmt1.Update(0, toHex("02"), []byte("v2"))
	require.Nil(t, err)
	rootHash1 := jmt1.Commit()

	// transit from v0 to v1
	jmt11, err := New(rootHash1, s1)
	require.Nil(t, err)
	err = jmt11.Update(1, toHex("02"), []byte("v3"))
	require.Nil(t, err)
	//printJMT(jmt11, 1)
	err = jmt11.Update(1, toHex("04"), []byte("v4"))
	require.Nil(t, err)
	rootHash11 := jmt11.Commit()

	// verify v0
	jmt1, err = New(rootHash1, s1)
	require.Nil(t, err)
	n, err := jmt1.Get(toHex("02"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))

	// verify v1
	jmt11, err = New(rootHash11, s1)
	require.Nil(t, err)
	n, err = jmt11.Get(toHex("02"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt11.Get(toHex("04"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))

	// ===========

	// init version 0 jmt
	jmt2, s2 := initEmptyJMT()
	err = jmt2.Update(0, toHex("02"), []byte("v2"))
	require.Nil(t, err)
	rootHash2 := jmt2.Commit()

	// transit from v0 to v1, but with different update order
	jmt22, err := New(rootHash2, s2)
	require.Nil(t, err)
	err = jmt22.Update(1, toHex("04"), []byte("v4"))
	require.Nil(t, err)
	err = jmt22.Update(1, toHex("02"), []byte("v3"))
	require.Nil(t, err)
	rootHash22 := jmt22.Commit()

	// verify v0
	jmt2, err = New(rootHash2, s2)
	require.Nil(t, err)
	n, err = jmt2.Get(toHex("02"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))

	// verify v1
	jmt22, err = New(rootHash22, s2)
	require.Nil(t, err)
	n, err = jmt22.Get(toHex("02"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))
	n, err = jmt22.Get(toHex("04"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v4"))

	require.Equal(t, rootHash11, rootHash22)
}

func Test_StateTransitWithDifferentDeleteOrder(t *testing.T) {
	// init version 0 jmt
	jmt1, s1 := initEmptyJMT()
	err := jmt1.Update(0, toHex("01"), []byte("v1"))
	require.Nil(t, err)
	err = jmt1.Update(0, toHex("02"), []byte("v2"))
	require.Nil(t, err)
	err = jmt1.Update(0, toHex("03"), []byte("v3"))
	require.Nil(t, err)
	err = jmt1.Update(0, toHex("04"), []byte("v4"))
	require.Nil(t, err)
	rootHash1 := jmt1.Commit()

	// transit from v0 to v1
	jmt11, err := New(rootHash1, s1)
	require.Nil(t, err)
	err = jmt11.Update(1, toHex("02"), []byte{})
	require.Nil(t, err)
	//printJMT(jmt11, 1)
	err = jmt11.Update(1, toHex("04"), []byte{})
	require.Nil(t, err)
	rootHash11 := jmt11.Commit()

	// verify v0
	jmt1, err = New(rootHash1, s1)
	require.Nil(t, err)
	n, err := jmt1.Get(toHex("02"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v2"))

	// verify v1
	jmt11, err = New(rootHash11, s1)
	require.Nil(t, err)
	n, err = jmt11.Get(toHex("02"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt11.Get(toHex("03"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))

	// ===========

	// init version 0 jmt
	jmt2, s2 := initEmptyJMT()
	err = jmt2.Update(0, toHex("01"), []byte("v1"))
	require.Nil(t, err)
	err = jmt2.Update(0, toHex("02"), []byte("v2"))
	require.Nil(t, err)
	err = jmt2.Update(0, toHex("03"), []byte("v3"))
	require.Nil(t, err)
	err = jmt2.Update(0, toHex("04"), []byte("v4"))
	require.Nil(t, err)
	rootHash2 := jmt2.Commit()

	// transit from v0 to v1, but with different delete order
	jmt22, err := New(rootHash2, s2)
	require.Nil(t, err)
	err = jmt22.Update(1, toHex("04"), []byte{})
	require.Nil(t, err)
	err = jmt22.Update(1, toHex("02"), []byte{})
	require.Nil(t, err)
	rootHash22 := jmt22.Commit()

	// verify v0
	jmt2, err = New(rootHash2, s2)
	require.Nil(t, err)
	n, err = jmt2.Get(toHex("03"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))

	// verify v1
	jmt22, err = New(rootHash22, s2)
	require.Nil(t, err)
	n, err = jmt22.Get(toHex("02"))
	require.Nil(t, err)
	require.Nil(t, n)
	n, err = jmt22.Get(toHex("03"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v3"))

	require.Equal(t, rootHash11, rootHash22)
}

func Test_Case_Random_1(t *testing.T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	maxn := 1000 // max number of nodes
	cnt := 10    // number of test cases
	for i := 0; i < cnt; i++ {
		jmt, s := initEmptyJMT()
		// nnum := rand.Intn(maxn)
		nnum := maxn
		fmt.Println("【Random Testcase ", i, "】, node num:", nnum)
		kv := make(map[string][]byte, nnum)
		for j := 0; j < nnum; j++ {
			k, v := getRandomHexKV(4, 16)
			kv[string(k)] = v
			err := jmt.Update(0, k, v)
			require.Nil(t, err)
		}
		rootHash := jmt.Commit()
		jmt, err := New(rootHash, s)
		require.Nil(t, err)
		for k, v := range kv {
			n, err := jmt.Get(([]byte)(k))
			require.Nil(t, err)
			require.NotNil(t, n)
			require.Equal(t, n, v)
		}
	}
}

func Test_Case_Random_2(t *testing.T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	maxn := 5000 // max number of nodes
	cases := 1   // number of test cases
	version := 2 // number of states
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
			rootHash = jmt.Commit()
			v2hash[ver] = rootHash
			jmt = nil
		}
		// get all
		for ver := 0; ver < version; ver++ {
			jmt, err = New(v2hash[ver], s)
			require.Nil(t, err)
			for k, v := range v2inserted[ver] {
				n, err := jmt.Get(([]byte)(k))
				require.Nil(t, err)
				require.NotNil(t, n)
				require.Equal(t, n, v)
			}
			for k := range v2deleted[ver] {
				n, err := jmt.Get(([]byte)(k))
				require.Nil(t, err)
				require.Nil(t, n)
			}
		}
	}
}

func Test_Case_Random_3(t *testing.T) {
	rand.Seed(uint64(time.Now().UnixNano()))
	testcasesNum := 10
	dummyValue := []byte{9}
	for tn := 0; tn < testcasesNum; tn++ {
		fmt.Println("【Random Testcase】 ", tn)
		nnum := 10 // max number of nodes
		cnt := 10  // number of shuffle times
		rootHashs := make([]common.Hash, cnt)
		jmts := make([]*JMT, cnt)
		updateOrder := make([][][]byte, cnt)
		lk := 2
		lv := 1

		ks0, _ := getRandomHexKVSet(lk, lv, nnum)
		ks, vs := getRandomHexKVSet(lk, lv, nnum)
		var nodes0 []*TraversedNode

		for i := 0; i < cnt; i++ {
			// init base jmt
			jmt, _ := initEmptyJMT()
			for idx := 0; idx < len(ks0); idx++ {
				err := jmt.Update(1, ks0[idx], dummyValue)
				require.Nil(t, err)
			}
			jmt.Commit()
			//if len(nodes0) == 0 {
			//	nodes0 = jmt.Traverse(1)
			//}

			rand.Shuffle(len(ks), func(i, j int) {
				ks[i], ks[j] = ks[j], ks[i]
				vs[i], vs[j] = vs[j], vs[i]
			})

			for idx := 0; idx < len(ks); idx++ {
				err := jmt.Update(2, ks[idx], vs[idx])
				require.Nil(t, err)
				updateOrder[i] = append(updateOrder[i], ks[idx])
			}

			rootHash := jmt.Commit()

			rootHashs[i] = rootHash
			jmts[i] = jmt
		}

		// only error case
		for i := 0; i < len(rootHashs); i++ {
			if rootHashs[0] != rootHashs[i] {
				fmt.Println("=========[ERROR] jmt forked=========")
				fmt.Println("=========Traverse base jmt=========")
				for j := 0; j < len(nodes0); j++ {
					fmt.Printf("Node[%v]: %v\n", convertHex((*nodes0[j]).Path), (*(*nodes0[j]).Origin).Print())
				}
				fmt.Println("=========END Traverse base jmt=========")

				fmt.Println("=========Traverse jmt[0]=========")
				//printJMT(jmts[0], 0)
				fmt.Println("Update key order of jmt[0]")
				for j := 0; j < len(updateOrder[0]); j++ {
					fmt.Println(convertHex(updateOrder[0][j]))
				}

				fmt.Printf("=========Traverse jmt[%v]=========\n", i)
				//printJMT(jmts[i], 0)

				fmt.Printf("Update key order of jmt[%v]\n", i)
				for j := 0; j < len(updateOrder[i]); j++ {
					fmt.Println(convertHex(updateOrder[i][j]))
				}
			}
			require.Equal(t, rootHashs[0], rootHashs[i])
		}
	}
}

//func printJMT(jmt *JMT, version uint64) {
//	fmt.Print("======Start Print JMT========\n")
//	nodes := jmt.Traverse(version)
//	for j := 0; j < len(nodes); j++ {
//		fmt.Printf("Node[%v]: %v\n", convertHex((*nodes[j]).Path), (*(*nodes[j]).Origin).Print())
//	}
//	fmt.Print("======End Print JMT========\n")
//}

func convertHex(in []byte) string {
	hexString := "0123456789abcdef"
	var ret string
	for i := 0; i < len(in); i++ {
		ret += string(hexString[in[i]])
	}
	return ret
}

func initEmptyJMT() (*JMT, storage.Storage) {
	dir, _ := os.MkdirTemp("", "TestKV")
	s, _ := pebble.New(dir, nil, nil)
	// init dummy jmt
	rootHash := common.Hash{}
	rootNodeKey := &NodeKey{
		Version: 0,
		Path:    []byte{},
		Type:    []byte{},
	}
	nk := rootNodeKey.encode()
	s.Put(nk, nil)
	s.Put(rootHash[:], nk)
	jmt, _ := New(rootHash, s)
	return jmt, s
}

func getRandomHexKVSet(lk, lv, num int) ([][]byte, [][]byte) {
	rand.Seed(uint64(time.Now().UnixNano()))
	keySet := map[string]struct{}{}
	var ks [][]byte
	var vs [][]byte
	for i := 0; i < num; i++ {
		k := make([]byte, lk)
		v := make([]byte, lv)
		for j := 0; j < lk; j++ {
			k[j] = byte(rand.Intn(16))
		}
		for j := 0; j < lv; j++ {
			v[j] = byte(rand.Intn(16))
		}
		if _, ok := keySet[string(k)]; ok {
			continue
		}
		ks = append(ks, k)
		vs = append(vs, v)
		keySet[string(k)] = struct{}{}
	}
	return ks, vs
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

func initKV() storage.Storage {
	dir, _ := os.MkdirTemp("", "TestKV")
	s, _ := pebble.New(dir, nil, nil)
	// init dummy jmt
	rootHash := placeHolder
	rootNodeKey := NodeKey{
		Version: 0,
		Path:    []byte{},
		Type:    []byte{},
	}
	nk := rootNodeKey.encode()
	s.Put(nk, nil)
	s.Put(rootHash[:], nk)
	return s
}

func toHex(s string) []byte {
	k := hexutil.EncodeToNibbles(s)
	return k
}
