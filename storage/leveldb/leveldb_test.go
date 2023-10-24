package leveldb

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/axiomesh/axiom-kit/storage"
)

func TestIter_Next(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestNext")
	require.Nil(t, err)

	_, err = New(dir, nil)
	require.Nil(t, err)
	_, err = New(dir, nil)
	require.EqualValues(t, "resource temporarily unavailable", err.Error())
}

func TestLdb_Put(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPut")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	err = s.Close()
	require.Nil(t, err)
}

func TestLdb_Delete(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestDelete")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	s.Delete([]byte("key"))
}

func TestLdb_Get(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestGet")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	v1 := s.Get([]byte("key"))
	assert.Equal(t, v1, []byte("value"))
	s.Delete([]byte("key"))
	v2 := s.Get([]byte("key"))
	assert.True(t, v2 == nil)
}

func TestLdb_GetPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestGetPanic")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Get([]byte("key"))
	assert.True(t, false)
}

func TestLdb_PutPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestPutPanic")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("key"))
	assert.True(t, false)
}

func TestLdb_DeletePanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestDeletePanic")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Delete([]byte("key"))
	assert.True(t, false)
}

func TestLdb_Has(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestHas")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	key := []byte("key")
	r1 := s.Has(key)
	assert.True(t, !r1)
	s.Put(key, []byte("value"))
	r2 := s.Has(key)
	assert.True(t, r2)
	s.Delete(key)
	r3 := s.Has(key)
	assert.True(t, !r3)
}

func TestLdb_NewBatch(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestNewBatch")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	batch.Delete([]byte("key10"))
	batch.Commit()

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := s.Get([]byte(key))
		assert.EqualValues(t, key, value)
	}
}

func TestLdb_CommitPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestDeletePanic")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	err = s.Close()
	require.Nil(t, err)

	batch.Commit()
	assert.True(t, false)
}

func TestLdb_Iterator(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestIterator")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	batch.Commit()

	iter := s.Iterator([]byte("key0"), []byte("key9"))
	i := 0
	for iter.Next() {
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Value())
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Key())
		i++
	}
	assert.EqualValues(t, i, 9)
}

func TestLdb_Iterator_Empty(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestIterator")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	batch.Commit()

	iter := s.Iterator([]byte("none"), []byte("no"))
	i := 0
	for iter.Next() {
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Value())
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Key())
		i++
	}
	assert.EqualValues(t, i, 0)
}

func TestLdb_Prefix(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPrefix")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	delKey := []byte("key11")
	batch.Delete(delKey)
	batch.Commit()

	iter := s.Prefix([]byte("key1"))
	expected := []string{"key1", "key10", "key12", "key13", "key14"}

	i := 0
	for iter.Next() {
		assert.EqualValues(t, []byte(expected[i]), iter.Value())
		assert.EqualValues(t, []byte(expected[i]), iter.Key())
		i++
	}
	assert.EqualValues(t, i, len(expected))
}

func TestLdb_Seek(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSeek")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	delKey := []byte("key5")
	batch.Delete(delKey)
	batch.Commit()

	iter := s.Iterator([]byte("key0"), []byte("key9"))
	assert.True(t, iter.Seek([]byte("key5")))

	expected := []string{"key7", "key8"}
	i := 0
	for iter.Next() {
		assert.EqualValues(t, []byte(expected[i]), iter.Value())
		assert.EqualValues(t, []byte(expected[i]), iter.Key())
		i++
	}
	assert.EqualValues(t, i, len(expected))
}

func TestLdb_Prev(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPrev")
	require.Nil(t, err)

	s, err := New(dir, nil)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	delKey := []byte("key3")
	batch.Delete(delKey)
	batch.Commit()

	iter := s.Iterator([]byte("key0"), []byte("key9"))
	iter.Seek([]byte("key6"))
	expected := []string{"key5", "key4", "key2", "key1", "key0"}
	i := 0
	for iter.Prev() {
		assert.EqualValues(t, []byte(expected[i]), iter.Value())
		assert.EqualValues(t, []byte(expected[i]), iter.Key())
		i++
	}
	assert.EqualValues(t, i, len(expected))
}

func BenchmarkLdb_Get(b *testing.B) {
	path, err := os.MkdirTemp("", "*")
	assert.Nil(b, err)

	ldb, err := New(path, nil)
	assert.Nil(b, err)

	val := make([]byte, 1024*1024*1)
	for k := 0; k < len(val); k++ {
		val[k] = byte(rand.Int63n(128))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("abc.%d.%d", i, j)
			ldb.Put([]byte(key), val)

			v := ldb.Get([]byte(key))
			assert.Equal(b, val, v)
		}

		iterator := ldb.Prefix([]byte("abc"))
		for iterator.Next() {
			ldb.Delete(iterator.Key())
		}
	}

	_ = ldb.Close()
}

func BenchmarkLevelDbSuite(b *testing.B) {
	path, err := os.MkdirTemp("", "*")
	assert.Nil(b, err)

	storage.BenchKvSuite(b, func() storage.Storage {
		db, err := New(path, &opt.Options{})
		if err != nil {
			b.Fatal(err)
		}
		return db
	})
}
