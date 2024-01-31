package pebble

import (
	"github.com/cockroachdb/pebble"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/axiomesh/axiom-kit/storage"
)

type pdb struct {
	db *pebble.DB
	wo *pebble.WriteOptions
}

// todo (zqr): use logger to record panic
// todo (zqr): add metrics for kv

func New(path string, opts *pebble.Options, wo *pebble.WriteOptions) (storage.Storage, error) {
	db, err := pebble.Open(path, opts)
	if err != nil {
		return nil, err
	}

	return &pdb{
		db: db,
		wo: wo,
	}, nil
}

func (p *pdb) Put(key, value []byte) {
	if err := p.db.Set(key, value, p.wo); err != nil {
		panic(err)
	}
}

func (p *pdb) Delete(key []byte) {
	if err := p.db.Delete(key, p.wo); err != nil {
		panic(err)
	}
}

func (p *pdb) Get(key []byte) []byte {
	val, closer, err := p.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil
		}
		panic(err)
	}
	ret := make([]byte, len(val))
	copy(ret, val)
	closer.Close()
	return ret
}

func (p *pdb) Has(key []byte) bool {
	_, closer, err := p.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return false
		}
		panic(err)
	}
	closer.Close()
	return true
}

func (p *pdb) Iterator(start, end []byte) storage.Iterator {
	iter := &iter{
		iter: p.db.NewIter(&pebble.IterOptions{
			LowerBound: start,
			UpperBound: end,
		}),
		positioned: false,
	}
	iter.iter.First()
	return iter
}

func (p *pdb) Prefix(prefix []byte) storage.Iterator {
	ran := util.BytesPrefix(prefix)
	iter := &iter{
		iter: p.db.NewIter(&pebble.IterOptions{
			LowerBound: ran.Start,
			UpperBound: ran.Limit,
		}),
		positioned: false,
	}
	iter.iter.First()
	return iter
}

func (p *pdb) NewBatch() storage.Batch {
	return &pdbBatch{
		batch: p.db.NewBatch(),
		wo:    p.wo,
	}
}

func (p *pdb) Close() error {
	return p.db.Close()
}

type iter struct {
	iter       *pebble.Iterator
	positioned bool
}

func (it *iter) Prev() bool {
	return it.iter.Prev()
}

func (it *iter) Seek(key []byte) bool {
	k := make([]byte, len(key))
	copy(k, key)
	it.positioned = true
	return it.iter.SeekGE(k)
}

func (it *iter) Next() bool {
	if !it.iter.Valid() {
		return false
	}
	if !it.positioned {
		it.positioned = true
		return true
	}
	return it.iter.Next()
}

func (it *iter) Key() []byte {
	key := it.iter.Key()
	ret := make([]byte, len(key))
	copy(ret, key)
	return ret
}

func (it *iter) Value() []byte {
	val, err := it.iter.ValueAndErr()
	if err != nil {
		panic(err)
	}
	return val
}

type pdbBatch struct {
	batch *pebble.Batch
	wo    *pebble.WriteOptions
	size  int
}

func (p *pdbBatch) Put(key, value []byte) {
	p.batch.Set(key, value, nil)
	p.size += len(key) + len(value)
}

func (p *pdbBatch) Delete(key []byte) {
	p.batch.Delete(key, nil)
	p.size += len(key)
}

func (p *pdbBatch) Commit() {
	if err := p.batch.Commit(p.wo); err != nil {
		panic(err)
	}
}

func (p *pdbBatch) Size() int {
	return p.size
}

func (p *pdbBatch) Reset() {
	p.batch.Reset()
	p.size = 0
}
