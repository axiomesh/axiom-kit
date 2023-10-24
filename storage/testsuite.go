package storage

import (
	"bytes"
	"crypto/rand"
	"sort"
	"testing"
)

// BenchKvSuite runs a suite of benchmarks against a KV backend implementation.
func BenchKvSuite(b *testing.B, New func() Storage) {
	var (
		keys, vals   = makeDataset(1_000_000, 32, 32, false)
		sKeys, sVals = makeDataset(1_000_000, 32, 32, true)
	)

	// Run benchmarks sequentially
	b.Run("Write", func(b *testing.B) {
		benchWrite := func(b *testing.B, keys, vals [][]byte) {
			b.ResetTimer()
			b.ReportAllocs()

			db := New()
			defer db.Close()

			for i := 0; i < len(keys); i++ {
				db.Put(keys[i], vals[i])
			}
		}
		b.Run("WriteSorted", func(b *testing.B) {
			benchWrite(b, sKeys, sVals)
		})

		b.Run("WriteRandom", func(b *testing.B) {
			benchWrite(b, keys, vals)
		})
	})

	b.Run("BatchWrite", func(b *testing.B) {
		benchBatchWrite := func(b *testing.B, keys, vals [][]byte) {
			b.ResetTimer()
			b.ReportAllocs()

			db := New()
			defer db.Close()

			batch := db.NewBatch()
			for i := 0; i < len(keys); i++ {
				batch.Put(keys[i], vals[i])
			}
			batch.Commit()
		}
		b.Run("BenchWriteSorted", func(b *testing.B) {
			benchBatchWrite(b, sKeys, sVals)
		})
		b.Run("BenchWriteRandom", func(b *testing.B) {
			benchBatchWrite(b, keys, vals)
		})
	})

	b.Run("Read", func(b *testing.B) {
		benchRead := func(b *testing.B, keys, vals [][]byte) {
			db := New()
			defer db.Close()

			batch := db.NewBatch()
			for i := 0; i < len(keys); i++ {
				batch.Put(keys[i], vals[i])
			}
			batch.Commit()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < len(keys); i++ {
				db.Get(keys[i])
			}
		}
		b.Run("ReadSorted", func(b *testing.B) {
			benchRead(b, sKeys, sVals)
		})

		b.Run("ReadRandom", func(b *testing.B) {
			benchRead(b, keys, vals)
		})
	})
}

func makeDataset(size, ksize, vsize int, order bool) ([][]byte, [][]byte) {
	var keys [][]byte
	var vals [][]byte
	for i := 0; i < size; i += 1 {
		keys = append(keys, randBytes(ksize))
		vals = append(vals, randBytes(vsize))
	}

	// order generated slice according to bytes order
	if order {
		sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i], keys[j]) < 0 })
	}
	return keys, vals
}

// randomHash generates a random blob of data and returns it as a hash.
func randBytes(len int) []byte {
	buf := make([]byte, len)
	if n, err := rand.Read(buf); n != len || err != nil {
		panic(err)
	}
	return buf
}
