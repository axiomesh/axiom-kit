package leveldb

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	leveldberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/axiomesh/axiom-kit/storage/kv"
)

type ldb struct {
	db *leveldb.DB
}

func New(path string, o *opt.Options) (kv.Storage, error) {
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, err
	}

	return &ldb{
		db: db,
	}, nil
}

func (l *ldb) Put(key, value []byte) {
	if err := l.db.Put(key, value, nil); err != nil {
		panic(err)
	}
}

func (l *ldb) Delete(key []byte) {
	if err := l.db.Delete(key, nil); err != nil {
		panic(err)
	}
}

func (l *ldb) Get(key []byte) []byte {
	val, err := l.db.Get(key, nil)
	if err != nil {
		if errors.Is(err, leveldberrors.ErrNotFound) {
			return nil
		}
		panic(err)
	}
	return val
}

func (l *ldb) Has(key []byte) bool {
	return l.Get(key) != nil
}

func (l *ldb) Iterator(start, end []byte) kv.Iterator {
	rg := &util.Range{
		Start: start,
		Limit: end,
	}
	it := l.db.NewIterator(rg, nil)

	return &iter{iter: it}
}

func (l *ldb) Prefix(prefix []byte) kv.Iterator {
	rg := util.BytesPrefix(prefix)

	return &iter{iter: l.db.NewIterator(rg, nil)}
}

func (l *ldb) NewBatch() kv.Batch {
	return &ldbBatch{
		ldb:   l.db,
		batch: &leveldb.Batch{},
	}
}

func (l *ldb) Close() error {
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
