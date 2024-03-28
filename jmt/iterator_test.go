package jmt

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/pebble"
)

func Test_IterateEmptyTree(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	jmt.Commit(nil)

	iter := NewIterator(jmt.root.GetHash(), s, nil, 100, time.Second)
	go iter.Iterate()
	var res []*RawNode

	for {
		node, err := iter.Next()
		if err != nil {
			if err == ErrorNoMoreData {
				require.Equal(t, 1, len(res))
				break
			} else {
				panic(err)
			}
		} else {
			res = append(res, node)
		}
	}
}

func Test_IterateStop(t *testing.T) {
	for tc := 0; tc < 10; tc++ {
		jmt, s := initEmptyJMT()
		err := jmt.Update(0, toHex("0001"), []byte("v1"))
		require.Nil(t, err)
		err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
		require.Nil(t, err)
		rootHash0 := jmt.Commit(nil)

		var res []*RawNode
		iter := NewIterator(rootHash0, s, nil, 1, 1000*time.Millisecond)
		go iter.Iterate()

		time.Sleep(100 * time.Millisecond)

		iter.Stop()
		for {
			n, err := iter.Next()
			if err != nil {
				require.Nil(t, n)
				if err != ErrorNoMoreData {
					require.Equal(t, err, ErrorInterrupted)
				} else {
					require.Equal(t, 1, len(res))
				}
				break
			}
			require.NotNil(t, n)
			res = append(res, n)
		}
	}
}

func Test_IterateTimeout(t *testing.T) {
	for tc := 0; tc < 10; tc++ {
		jmt, s := initEmptyJMT()
		err := jmt.Update(0, toHex("0001"), []byte("v1"))
		require.Nil(t, err)
		err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
		require.Nil(t, err)
		rootHash0 := jmt.Commit(nil)

		var res []*RawNode
		iter := NewIterator(rootHash0, s, nil, 1, 100*time.Millisecond)
		go iter.Iterate()

		time.Sleep(500 * time.Millisecond)

		for {
			n, err := iter.Next()
			if err != nil {
				require.Nil(t, n)
				if err != ErrorNoMoreData {
					require.Equal(t, err, ErrorTimeout)
				} else {
					require.Equal(t, 1, len(res))
				}
				break
			} else {
				require.NotNil(t, n)
				res = append(res, n)
			}
		}
	}
}

//		                      [_]
//	                      /         \
//					   [0]          [b]
//			             |              |
//	                  [00]          [bb]
//	                    |              |
//	               [000]            ——————
//	                   ｜           |       |
//		   	       ——————————    <bb17>   <bbf7>
//		           |       |
//			    <0001>  <0003>
func Test_IterateHistoryTrie(t *testing.T) {
	for tc := 0; tc < 10; tc++ {
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
		rootHash0 := jmt.Commit(nil)
		require.Equal(t, rootHash0, jmt.root.GetHash())
		jmt, err = New(rootHash0, s, nil, nil, jmt.logger)
		require.Nil(t, err)
		// transit from v0 to v1
		err = jmt.Update(1, toHex("0001"), []byte("v5"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("bbf7"), []byte("v6"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("bb17"), []byte("v8"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("0011"), []byte("v11"))
		require.Nil(t, err)
		// commit version 1 jmt, and load it from kv
		rootHash1 := jmt.Commit(nil)
		require.Equal(t, rootHash1, jmt.root.GetHash())
		jmt, err = New(rootHash1, s, nil, nil, jmt.logger)
		require.Nil(t, err)
		// transit from v1 to v2
		err = jmt.Update(2, toHex("0001"), []byte("v9"))
		require.Nil(t, err)
		err = jmt.Update(2, toHex("bbf7"), []byte("v10"))
		require.Nil(t, err)
		err = jmt.Update(2, toHex("bb17"), []byte("v12"))
		require.Nil(t, err)
		rootHash2 := jmt.Commit(nil)
		require.Equal(t, rootHash2, jmt.root.GetHash())

		// iterate version 0 jmt trie
		iter := NewIterator(rootHash0, s, nil, 2, time.Second)
		go iter.Iterate()
		var res []*RawNode

		for {
			data, err := iter.Next()
			if err != nil {
				require.Nil(t, data)
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		root0NK := s.Get(rootHash0[:])

		// persist v0 jmt trie
		s0 := initKVStorage()
		batch := s0.NewBatch()
		batch.Put(rootHash0[:], root0NK)
		for _, n := range res {
			batch.Put(n.RawKey, n.RawValue)
		}
		batch.Commit()

		// verify v0 trie
		verified, err := VerifyTrie(rootHash0, s0, nil)
		require.Nil(t, err)
		require.True(t, verified)

		// verify v0 state
		jmt, err = New(rootHash0, s0, nil, nil, jmt.logger)
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

		// iterate version 1 jmt trie
		iter = NewIterator(rootHash1, s, nil, 1, time.Second)
		go iter.Iterate()
		res = []*RawNode{}
		for {
			data, err := iter.Next()
			if err != nil {
				require.Nil(t, data)
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		root1NK := s.Get(rootHash1[:])

		// persist v1 jmt trie
		s1 := initKVStorage()
		batch = s1.NewBatch()
		batch.Put(rootHash1[:], root1NK)
		for _, n := range res {
			batch.Put(n.RawKey, n.RawValue)
		}
		batch.Commit()

		// verify v1 trie
		verified, err = VerifyTrie(rootHash1, s1, nil)
		require.Nil(t, err)
		require.True(t, verified)

		// verify v1 state
		jmt, err = New(rootHash1, s1, nil, nil, jmt.logger)
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
		n, err = jmt.Get(toHex("0011"))
		require.Nil(t, err)
		require.NotNil(t, n)
		require.Equal(t, n, []byte("v11"))

		// iterate version 2 jmt trie
		iter = NewIterator(rootHash2, s, nil, 10, time.Second)
		go iter.Iterate()
		res = []*RawNode{}
		for {
			data, err := iter.Next()
			if err != nil {
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		root2NK := s.Get(rootHash2[:])

		// persist v2 jmt trie
		s2 := initKVStorage()
		batch = s2.NewBatch()
		batch.Put(rootHash2[:], root2NK)
		for _, n := range res {
			batch.Put(n.RawKey, n.RawValue)
		}
		batch.Commit()

		// verify v2 trie
		verified, err = VerifyTrie(rootHash2, s2, nil)
		require.Nil(t, err)
		require.True(t, verified)

		// verify v2 state
		jmt, err = New(rootHash2, s2, nil, nil, jmt.logger)
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
		n, err = jmt.Get(toHex("0011"))
		require.Nil(t, err)
		require.NotNil(t, n)
		require.Equal(t, n, []byte("v11"))
	}
}

func Test_IterateHistoryTrieLeafOnly(t *testing.T) {
	for tc := 0; tc < 10; tc++ {
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
		leafSet1 := map[string][]byte{}
		leafSet1[string(toHex("0001"))] = []byte("v1")
		leafSet1[string(toHex("bbf7"))] = []byte("v2")
		leafSet1[string(toHex("0003"))] = []byte("v3")
		leafSet1[string(toHex("bb17"))] = []byte("v4")
		// commit version 0 jmt, and load it from kv
		rootHash0 := jmt.Commit(nil)
		require.Equal(t, rootHash0, jmt.root.GetHash())
		jmt, err = New(rootHash0, s, nil, nil, jmt.logger)
		require.Nil(t, err)

		// transit from v0 to v1
		err = jmt.Update(1, toHex("0001"), []byte("v5"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("bbf7"), []byte("v6"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("bb17"), []byte("v8"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("0011"), []byte("v11"))
		require.Nil(t, err)
		err = jmt.Update(1, toHex("abcd"), []byte("v1"))
		require.Nil(t, err)
		leafSet2 := map[string][]byte{}
		leafSet2[string(toHex("0001"))] = []byte("v5")
		leafSet2[string(toHex("bbf7"))] = []byte("v6")
		leafSet2[string(toHex("0011"))] = []byte("v11")
		leafSet2[string(toHex("bb17"))] = []byte("v8")
		leafSet2[string(toHex("0003"))] = []byte("v3")
		leafSet2[string(toHex("abcd"))] = []byte("v1")
		// commit version 1 jmt, and load it from kv
		rootHash1 := jmt.Commit(nil)
		require.Equal(t, rootHash1, jmt.root.GetHash())
		jmt, err = New(rootHash1, s, nil, nil, jmt.logger)
		require.Nil(t, err)

		// transit from v1 to v2
		err = jmt.Update(2, toHex("0001"), []byte("v9"))
		require.Nil(t, err)
		err = jmt.Update(2, toHex("bbf7"), []byte("v10"))
		require.Nil(t, err)
		err = jmt.Update(2, toHex("bb17"), []byte("v12"))
		require.Nil(t, err)
		err = jmt.Update(2, toHex("ac00"), []byte("v13"))
		require.Nil(t, err)
		leafSet3 := map[string][]byte{}
		leafSet3[string(toHex("0001"))] = []byte("v9")
		leafSet3[string(toHex("bbf7"))] = []byte("v10")
		leafSet3[string(toHex("0011"))] = []byte("v11")
		leafSet3[string(toHex("bb17"))] = []byte("v12")
		leafSet3[string(toHex("0003"))] = []byte("v3")
		leafSet3[string(toHex("abcd"))] = []byte("v1")
		leafSet3[string(toHex("ac00"))] = []byte("v13")
		rootHash2 := jmt.Commit(nil)
		require.Equal(t, rootHash2, jmt.root.GetHash())

		// iterate version 0 jmt trie leaf
		iter := NewIterator(rootHash0, s, nil, 2, time.Second)
		go iter.IterateLeaf()
		var res []*RawNode

		for {
			data, err := iter.Next()
			if err != nil {
				require.Nil(t, data)
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		require.Equal(t, len(res), len(leafSet1))
		for _, d := range res {
			require.True(t, len(d.LeafKey) != 0)
			require.Equal(t, d.LeafValue, leafSet1[string(d.LeafKey)])
		}

		// iterate version 1 jmt trie
		iter = NewIterator(rootHash1, s, nil, 1, time.Second)
		go iter.IterateLeaf()
		res = []*RawNode{}
		for {
			data, err := iter.Next()
			if err != nil {
				require.Nil(t, data)
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		require.Equal(t, len(res), len(leafSet2))
		for _, d := range res {
			require.True(t, len(d.LeafKey) != 0)
			require.Equal(t, d.LeafValue, leafSet2[string(d.LeafKey)])
		}

		// iterate version 2 jmt trie
		iter = NewIterator(rootHash2, s, nil, 10, time.Second)
		go iter.IterateLeaf()
		res = []*RawNode{}
		for {
			data, err := iter.Next()
			if err != nil {
				if err == ErrorNoMoreData {
					break
				}
				panic(err)
			}
			require.NotNil(t, data)
			res = append(res, data)
		}
		require.Equal(t, len(res), len(leafSet3))
		for _, d := range res {
			require.True(t, len(d.LeafKey) != 0)
			require.Equal(t, d.LeafValue, leafSet3[string(d.LeafKey)])
		}
	}
}

func initKVStorage() storage.Storage {
	dir, _ := os.MkdirTemp("", "TestKVStorage")
	s, _ := pebble.New(dir, nil, nil, logrus.New())
	return s
}
