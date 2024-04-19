package kv

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	leveldberrors "github.com/syndtr/goleveldb/leveldb/errors"
	leveldbstorage "github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type memory struct {
	db *leveldb.DB
}

func NewMemory() Storage {
	db, err := leveldb.Open(leveldbstorage.NewMemStorage(), nil)
	if err != nil {
		panic(errors.Errorf("failed to new memory storage for leveldb: %v", err))
	}
	return &memory{
		db: db,
	}
}

func (l *memory) Put(key, value []byte) {
	if err := l.db.Put(key, value, nil); err != nil {
		panic(err)
	}
}

func (l *memory) Delete(key []byte) {
	if err := l.db.Delete(key, nil); err != nil {
		panic(err)
	}
}

func (l *memory) Get(key []byte) []byte {
	val, err := l.db.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldberrors.ErrNotFound) {
			return nil
		}
		panic(err)
	}
	return val
}

func (l *memory) Has(key []byte) bool {
	return l.Get(key) != nil
}

func (l *memory) Iterator(start, end []byte) Iterator {
	return l.db.NewIterator(&util.Range{
		Start: start,
		Limit: end,
	}, nil)
}

func (l *memory) Prefix(prefix []byte) Iterator {
	return l.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (l *memory) NewBatch() Batch {
	return &ldbBatch{
		ldb:   l.db,
		batch: &leveldb.Batch{},
	}
}

func (l *memory) Close() error {
	return l.db.Close()
}

type ldbBatch struct {
	ldb   *leveldb.DB
	batch *leveldb.Batch
	size  int
}

func (l *ldbBatch) Put(key, value []byte) {
	l.batch.Put(key, value)
	l.size += len(key) + len(value)
}

func (l *ldbBatch) Delete(key []byte) {
	l.batch.Delete(key)
	l.size += len(key)
}

func (l *ldbBatch) Commit() {
	if err := l.ldb.Write(l.batch, nil); err != nil {
		panic(err)
	}
}

func (l *ldbBatch) Size() int {
	return l.size
}

func (l *ldbBatch) Reset() {
	l.batch.Reset()
	l.size = 0
}
