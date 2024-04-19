package pebble

import (
	"errors"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/axiomesh/axiom-kit/storage/kv"
)

const (
	metricsGatherInterval = time.Second
	mbSize                = 1000 * 1000
)

type pdb struct {
	db *pebble.DB
	wo *pebble.WriteOptions

	logger logrus.FieldLogger

	closed bool // keep track of whether we're Closed

	metrics *Metrics
}

func New(path string, opts *pebble.Options, wo *pebble.WriteOptions, logger logrus.FieldLogger, metricsOpts ...MetricsOption) (kv.Storage, error) {
	db, err := pebble.Open(path, opts)
	if err != nil {
		return nil, err
	}

	pebbleDB := &pdb{
		db:      db,
		wo:      wo,
		metrics: &Metrics{},
		logger:  logger,
	}

	for _, opt := range metricsOpts {
		opt(pebbleDB.metrics)
	}

	go pebbleDB.meter(metricsGatherInterval)

	return pebbleDB, nil
}

func (p *pdb) Put(key, value []byte) {
	if err := p.db.Set(key, value, p.wo); err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble put failed")
	}
}

func (p *pdb) Delete(key []byte) {
	if err := p.db.Delete(key, p.wo); err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble delete failed")
	}
}

func (p *pdb) Get(key []byte) []byte {
	val, closer, err := p.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil
		}
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble get failed")
		return nil
	}
	ret := make([]byte, len(val))
	copy(ret, val)
	if err := closer.Close(); err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Pebble get closer close failed")
	}
	return ret
}

func (p *pdb) Has(key []byte) bool {
	_, closer, err := p.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false
		}
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble judge key has failed")
		return false
	}
	if err := closer.Close(); err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Pebble has closer close failed")
	}
	return true
}

func (p *pdb) Iterator(start, end []byte) kv.Iterator {
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	})
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble NewIter failed")
		return nil
	}
	iter := &iter{
		iter:       it,
		positioned: false,
		logger:     p.logger,
	}
	iter.iter.First()
	return iter
}

func (p *pdb) Prefix(prefix []byte) kv.Iterator {
	ran := util.BytesPrefix(prefix)
	it, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: ran.Start,
		UpperBound: ran.Limit,
	})
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble NewIter failed")
		return nil
	}
	iter := &iter{
		iter:       it,
		positioned: false,
		logger:     p.logger,
	}
	iter.iter.First()
	return iter
}

func (p *pdb) NewBatch() kv.Batch {
	return &pdbBatch{
		batch:  p.db.NewBatch(),
		wo:     p.wo,
		logger: p.logger,
	}
}

func (p *pdb) Close() error {
	err := p.db.Close()
	if err != nil {
		return err
	}
	p.closed = true
	return nil
}

type iter struct {
	iter       *pebble.Iterator
	positioned bool
	logger     logrus.FieldLogger
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
		it.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble iter value failed")
		return nil
	}
	return val
}

type pdbBatch struct {
	batch  *pebble.Batch
	wo     *pebble.WriteOptions
	size   int
	logger logrus.FieldLogger
}

func (p *pdbBatch) Put(key, value []byte) {
	err := p.batch.Set(key, value, nil)
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Pebble batch set failed")
	}
	p.size += len(key) + len(value)
}

func (p *pdbBatch) Delete(key []byte) {
	err := p.batch.Delete(key, nil)
	if err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Pebble batch delete failed")
	}
	p.size += len(key)
}

func (p *pdbBatch) Commit() {
	if err := p.batch.Commit(p.wo); err != nil {
		p.logger.WithFields(logrus.Fields{
			"err": err,
		}).Panic("Pebble batch commit failed")
	}
}

func (p *pdbBatch) Size() int {
	return p.size
}

func (p *pdbBatch) Reset() {
	p.batch.Reset()
	p.size = 0
}

// meter periodically retrieves internal pebble metrics
func (p *pdb) meter(refresh time.Duration) {
	timer := time.NewTimer(refresh)
	defer timer.Stop()

	// Create storage and warning log tracer for write delay.
	var (
		nDiskWrites      [2]int64
		nWalWrites       [2]int64
		nEffectiveWrites [2]int64
	)

	// Iterate ad infinitum and collect the stats
	for i := 1; ; i++ {
		var (
			nDiskWrite      int64
			nWalWrite       int64
			nEffectiveWrite int64

			pebbleInternalMetrics = p.db.Metrics()
		)

		for _, levelMetrics := range pebbleInternalMetrics.Levels {
			nDiskWrite += int64(levelMetrics.BytesCompacted)
			nDiskWrite += int64(levelMetrics.BytesFlushed)
		}

		nDiskWrite += int64(pebbleInternalMetrics.WAL.BytesWritten)
		nDiskWrites[i%2] = nDiskWrite

		nWalWrite += int64(pebbleInternalMetrics.WAL.BytesWritten)
		nWalWrites[i%2] = nWalWrite

		nEffectiveWrite += int64(pebbleInternalMetrics.WAL.BytesIn)
		nEffectiveWrites[i%2] = nEffectiveWrite

		if p.metrics.diskSizeGauge != nil {
			p.metrics.diskSizeGauge.Set(float64(int64(pebbleInternalMetrics.DiskSpaceUsage())) / mbSize)
		}

		if p.metrics.diskWriteThroughput != nil {
			p.metrics.diskWriteThroughput.Set(float64(nDiskWrites[i%2]-nDiskWrites[(i-1)%2]) / mbSize)
		}

		if p.metrics.walWriteThroughput != nil {
			p.metrics.walWriteThroughput.Set(float64(nWalWrites[i%2]-nWalWrites[(i-1)%2]) / mbSize)
		}

		if p.metrics.effectiveWriteThroughput != nil {
			p.metrics.effectiveWriteThroughput.Set(float64(nEffectiveWrites[i%2]-nEffectiveWrites[(i-1)%2]) / mbSize)
		}

		if p.closed {
			return
		}
		<-timer.C
		timer.Reset(refresh)
	}
}
