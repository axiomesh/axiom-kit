package jmt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
)

//		                              [0_]
//	                          /                     \
//					        [0_0]                 [0_b]
//			                  |                     |
//	                       [0_00]                 [0_bb]
//	                         |
//	                     __________                   |
//	                     |         |               ——————
//	               [0_000]      <0_0044>         |         |
//	                   ｜                     <0_bb17>   <0_bbf7>
//		   	       ——————————
//		           |       |
//			    <0_0001>  <0_0003>
func Test_PruneJournal(t *testing.T) {
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

	rootHash0, journal := jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 10)
	require.Equal(t, len(journal.PruneSet), 0)
	prune(jmt.backend, journal) // persist basic jmt

	jmt, err = New(rootHash0, s, nil, jmt.logger)
	err = jmt.Update(1, toHex("0003"), nil)
	require.Nil(t, err)
	_, journal = jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 2)
	require.Equal(t, len(journal.PruneSet), 6)

	jmt, err = New(rootHash0, s, nil, jmt.logger)
	err = jmt.Update(1, toHex("bbf7"), nil)
	err = jmt.Update(1, toHex("bb17"), nil)
	require.Nil(t, err)
	_, journal = jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 1)
	require.Equal(t, len(journal.PruneSet), 5)

	jmt, err = New(rootHash0, s, nil, jmt.logger)
	err = jmt.Update(1, toHex("0003"), []byte("v5"))
	require.Nil(t, err)
	_, journal = jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 5)
	require.Equal(t, len(journal.PruneSet), 5)

	jmt, err = New(rootHash0, s, nil, jmt.logger)
	err = jmt.Update(1, toHex("0044"), []byte("v6"))
	require.Nil(t, err)
	rootHash1, journal := jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 4)
	require.Equal(t, len(journal.PruneSet), 3)
	prune(jmt.backend, journal) // persist v1 jmt

	jmt, err = New(rootHash1, s, nil, jmt.logger)
	err = jmt.Update(2, toHex("0001"), nil)
	require.Nil(t, err)
	_, journal = jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 4)
	require.Equal(t, len(journal.PruneSet), 6)

	jmt, err = New(rootHash1, s, nil, jmt.logger)
	err = jmt.Update(2, toHex("0044"), nil)
	require.Nil(t, err)
	_, journal = jmt.Commit(false)
	require.Equal(t, len(journal.DirtySet), 3)
	require.Equal(t, len(journal.PruneSet), 4)
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
func Test_PruneHistoryWithOnlyInsert(t *testing.T) {
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
	rootHash0, journal0 := jmt.Commit(false)
	require.Equal(t, rootHash0, jmt.root.GetHash())
	prune(jmt.backend, journal0)
	jmt, err = New(rootHash0, s, nil, jmt.logger)
	require.Nil(t, err)

	// verify v0
	jmt, err = New(rootHash0, s, nil, jmt.logger)
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

	// transit from v0 to v1
	err = jmt.Update(1, toHex("0001"), []byte("v5"))
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bbf7"), []byte("v6"))
	require.Nil(t, err)
	err = jmt.Update(1, toHex("bb17"), []byte("v8"))
	require.Nil(t, err)

	// commit version 1 jmt, and load it from kv
	rootHash1, journal1 := jmt.Commit(false)
	require.Equal(t, rootHash1, jmt.root.GetHash())
	prune(jmt.backend, journal1)
	jmt, err = New(rootHash1, s, nil, jmt.logger)
	require.Nil(t, err)

	// verify v1
	jmt, err = New(rootHash1, s, nil, jmt.logger)
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

	// transit from v1 to v2
	err = jmt.Update(2, toHex("0001"), []byte("v9"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bbf7"), []byte("v10"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bb17"), []byte("v12"))
	require.Nil(t, err)
	rootHash2, journal2 := jmt.Commit(false)
	require.Equal(t, rootHash2, jmt.root.GetHash())
	prune(jmt.backend, journal2)

	// verify v2
	jmt, err = New(rootHash2, s, nil, jmt.logger)
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

//		                              [0_]
//	                          /                     \
//					        [0_0]                 [0_b]
//			                  |                     |
//	                       [0_00]                 [0_bb]
//	                         |
//	                     __________                   |
//	                     |         |               ——————
//	               [0_000]      <0_0044>         |         |
//	                   ｜                     <0_bb17>   <0_bbf7>
//		   	       ——————————
//		           |       |
//			    <0_0001>  <0_0003>
func Test_PruneHistoryWithDelete(t *testing.T) {
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
	err = jmt.Update(0, toHex("0044"), []byte("v5"))
	require.Nil(t, err)
	// commit version 0 jmt, and load it from kv
	rootHash0, _ := jmt.Commit(true)
	require.Equal(t, rootHash0, jmt.root.GetHash())
	jmt, err = New(rootHash0, s, nil, jmt.logger)
	require.Nil(t, err)

	// verify v0
	jmt, err = New(rootHash0, s, nil, jmt.logger)
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
	n, err = jmt.Get(toHex("0044"))
	require.Nil(t, err)
	require.NotNil(t, n)
	require.Equal(t, n, []byte("v5"))

	// transit from v0 to v1
	err = jmt.Update(1, toHex("0044"), []byte{})
	require.Nil(t, err)

	// commit version 1 jmt, and load it from kv
	rootHash1, journal1 := jmt.Commit(false)
	require.Equal(t, rootHash1, jmt.root.GetHash())
	prune(jmt.backend, journal1)
	jmt, err = New(rootHash1, s, nil, jmt.logger)
	require.Nil(t, err)

	// verify v1
	jmt, err = New(rootHash1, s, nil, jmt.logger)
	require.Nil(t, err)
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
	n, err = jmt.Get(toHex("0044"))
	require.Nil(t, err)
	require.Nil(t, n)

	// transit from v1 to v2
	err = jmt.Update(2, toHex("0001"), []byte("v7"))
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bbf7"), []byte{})
	require.Nil(t, err)
	err = jmt.Update(2, toHex("bb17"), []byte("v8"))
	require.Nil(t, err)
	rootHash2, journal2 := jmt.Commit(false)
	require.Equal(t, rootHash2, jmt.root.GetHash())
	prune(jmt.backend, journal2)

	// verify v2
	jmt, err = New(rootHash2, s, nil, jmt.logger)
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

func prune(backend storage.Storage, journal *types.TrieJournal) {
	batch := backend.NewBatch()
	for k := range journal.PruneSet {
		batch.Delete([]byte(k))
	}
	for k, v := range journal.DirtySet {
		batch.Put([]byte(k), v)
	}
	batch.Commit()
}
