package jmt

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/storage/pebble"
)

func Test_IterateEmptyTree(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	jmt.Commit()

	iter := NewIterator(jmt.root.GetHash(), s, 100, time.Second)
	go iter.Iterate()
	var res []*RawNode

	for {
		select {
		case err, ok := <-iter.ErrC:
			if ok {
				require.Nil(t, err)
			}
		case data, ok := <-iter.BufferC:

			if ok {
				//fmt.Printf("%v\n", data.print())
				res = append(res, data)
			}
		case <-iter.StopC:
			require.Equal(t, 1, len(res))
			return
		}
	}
}

func Test_IterateStop(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	rootHash0 := jmt.Commit()

	iter := NewIterator(rootHash0, s, 1, 1000*time.Millisecond)
	go iter.Iterate()

	time.Sleep(100 * time.Millisecond)

	iter.Stop()

	<-iter.StopC
	err, ok := <-iter.ErrC
	require.True(t, ok)
	require.NotNil(t, err)
	require.Equal(t, err, ErrorInterrupted)
}

func Test_IterateTimeout(t *testing.T) {
	jmt, s := initEmptyJMT()
	err := jmt.Update(0, toHex("0001"), []byte("v1"))
	require.Nil(t, err)
	err = jmt.Update(0, toHex("bbf7"), []byte("v2"))
	require.Nil(t, err)
	rootHash0 := jmt.Commit()

	iter := NewIterator(rootHash0, s, 1, 100*time.Millisecond)
	go iter.Iterate()

	time.Sleep(500 * time.Millisecond)

	<-iter.StopC
	err, ok := <-iter.ErrC
	require.True(t, ok)
	require.NotNil(t, err)
	require.Equal(t, err, ErrorTimeout)
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
	err = jmt.Update(1, toHex("0011"), []byte("v11"))
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

	// iterate version 0 jmt trie
	iter := NewIterator(rootHash0, s, 2, time.Second)
	go iter.Iterate()
	var res []*RawNode

	var finish bool
	for {
		select {
		case err, ok := <-iter.ErrC:
			if ok {
				require.Nil(t, err)
			}
		case <-iter.StopC:
			require.Equal(t, 10, len(res))
			finish = true
		default:
			for {
				data, ok := <-iter.BufferC
				if !ok {
					break
				}
				res = append(res, data)
			}
		}
		if finish {
			break
		}
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

	// verify v0
	jmt, err = New(rootHash0, s0)
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
	iter = NewIterator(rootHash1, s, 1, time.Second)
	go iter.Iterate()
	res = []*RawNode{}
	finish = false

	for {
		select {
		case err, ok := <-iter.ErrC:
			if ok {
				require.Nil(t, err)
			}
		case <-iter.StopC:
			require.Equal(t, 11, len(res))
			finish = true
		default:
			for {
				data, ok := <-iter.BufferC
				if !ok {
					break
				}
				res = append(res, data)
			}
		}
		if finish {
			break
		}
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

	// verify v1
	jmt, err = New(rootHash1, s1)
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
	iter = NewIterator(rootHash2, s, 10, time.Second)
	go iter.Iterate()
	res = []*RawNode{}
	finish = false

	for {
		select {
		case err, ok := <-iter.ErrC:
			if ok {
				require.Nil(t, err)
			}
		case <-iter.StopC:
			require.Equal(t, 11, len(res))
			finish = true
		default:
			for {
				data, ok := <-iter.BufferC
				if !ok {
					break
				}
				res = append(res, data)
			}
		}
		if finish {
			break
		}
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

	// verify v2
	jmt, err = New(rootHash2, s2)
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

func initKVStorage() storage.Storage {
	dir, _ := os.MkdirTemp("", "TestKVStorage")
	s, _ := pebble.New(dir, nil, nil)
	return s
}
