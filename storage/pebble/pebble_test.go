package pebble

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/storage"
)

var testLogger = log.NewWithModule("storage_test")

func TestIter_Next(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestNext")
	require.Nil(t, err)

	_, err = New(dir, nil, nil, testLogger)
	require.Nil(t, err)
	_, err = New(dir, nil, nil, testLogger)
	require.EqualValues(t, "lock held by current process", err.Error())
}

func TestPdb_Put(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPut")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	err = s.Close()
	require.Nil(t, err)
}

func TestPdb_Delete(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestDelete")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	s.Delete([]byte("key"))
}

func TestPdb_Get(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestGet")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("value"))
	v1 := s.Get([]byte("key"))
	assert.Equal(t, v1, []byte("value"))
	s.Delete([]byte("key"))
	v2 := s.Get([]byte("key"))
	assert.True(t, v2 == nil)
}

func TestPdb_GetPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestGetPanic")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Get([]byte("key"))
	assert.True(t, false)
}

func TestPdb_PutPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestPutPanic")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Put([]byte("key"), []byte("key"))
	assert.True(t, false)
}

func TestPdb_DeletePanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestDeletePanic")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	err = s.Close()
	require.Nil(t, err)

	s.Delete([]byte("key"))
	assert.True(t, false)
}

func TestPdb_Has(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestHas")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_NewBatch(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestNewBatch")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	deleteKey := "key5"
	batch.Delete([]byte(deleteKey))
	batch.Commit()

	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		value := s.Get([]byte(key))
		if key == deleteKey {
			assert.Nil(t, value)
		} else {
			assert.EqualValues(t, key, value)
		}
	}
}

func TestPdb_CommitPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.NotNil(t, err)
		}
	}()

	dir, err := os.MkdirTemp("", "TestDeletePanic")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_Iterator(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestIterator")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	batch := s.NewBatch()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
	}
	delKey := "key5"
	batch.Delete([]byte(delKey))
	batch.Commit()

	iter := s.Iterator([]byte("key0"), []byte("key9"))

	i, cnt := 0, 0
	for iter.Next() {
		if i == 5 {
			i++
		}
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Value())
		assert.EqualValues(t, []byte(fmt.Sprintf("key%d", i)), iter.Key())
		i++
		cnt++
	}
	assert.EqualValues(t, i, 9)
	assert.EqualValues(t, cnt, 8)
}

func TestPdb_Iterator_Empty(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestIterator")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_Prefix(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPrefix")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_Seek(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestSeek")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_Prev(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestPrev")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
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

func TestPdb_BatchSize(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestBatchSize")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	batch := s.NewBatch()
	total := 0
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
		total += 2 * len([]byte(key))
	}
	deleteKey := "key5"
	batch.Delete([]byte(deleteKey))
	total += len([]byte(deleteKey))
	assert.EqualValues(t, total, batch.Size())
	batch.Commit()

	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		value := s.Get([]byte(key))
		if key == deleteKey {
			assert.Nil(t, value)
		} else {
			assert.EqualValues(t, key, value)
		}
	}
}

func TestPdb_BatchReset(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestBatchReset")
	require.Nil(t, err)

	s, err := New(dir, nil, nil, testLogger)
	require.Nil(t, err)

	batch := s.NewBatch()
	total := 0
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
		total += 2 * len([]byte(key))
	}
	deleteKey := "key5"
	batch.Delete([]byte(deleteKey))
	total += len([]byte(deleteKey))
	assert.EqualValues(t, total, batch.Size())
	batch.Commit()

	batch.Reset()
	total = 0

	for i := 11; i < 20; i++ {
		key := fmt.Sprintf("key%d", i)
		batch.Put([]byte(key), []byte(key))
		total += 2 * len([]byte(key))
	}
	assert.EqualValues(t, total, batch.Size())
	batch.Commit()

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("key%d", i)
		value := s.Get([]byte(key))
		if key == deleteKey {
			assert.Nil(t, value)
		} else {
			assert.EqualValues(t, key, value)
		}
	}
}

func BenchmarkPebbleSuite(b *testing.B) {
	// Two memory tables is configured which is identical to leveldb,
	// including a frozen memory table and another live one.
	memTableLimit := 2
	cache := 256
	memTableSize := cache * 1024 * 1024 / 2 / memTableLimit

	opts := &pebble.Options{
		// Pebble has a single combined cache area and the write
		// buffers are taken from this too. Assign all available
		// memory allowance for cache.
		Cache: pebble.NewCache(int64(cache * 1024 * 1024)),

		// MaxOpenFiles: 1000,

		// The size of memory table(as well as the write buffer).
		// Note, there may have more than two memory tables in the system.
		MemTableSize: memTableSize,

		// MemTableStopWritesThreshold places a hard limit on the size
		// of the existent MemTables(including the frozen one).
		// Note, this must be the number of tables not the size of all memtables.
		MemTableStopWritesThreshold: memTableLimit,

		// The default compaction concurrency(1 thread),
		// Here use all available CPUs for faster compaction.
		MaxConcurrentCompactions: func() int { return runtime.NumCPU() },

		// Per-level options. Options for at least one level must be specified. The
		// options for the last level are used for all subsequent levels.
		Levels: []pebble.LevelOptions{
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
			{TargetFileSize: 2 * 1024 * 1024, FilterPolicy: bloom.FilterPolicy(10)},
		},
	}

	path, err := os.MkdirTemp("", "*")
	assert.Nil(b, err)

	storage.BenchKvSuite(b, func() storage.Storage {
		db, err := New(path, opts, nil, testLogger)
		if err != nil {
			b.Fatal(err)
		}
		return db
	})
}
