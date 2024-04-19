package blockfile

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/sirupsen/logrus"
)

// BlockFile block number start from 0
//
//go:generate mockgen -destination mock_blockfile/mock_blockfile.go -package mock_blockfile -source blockfile.go -typed
type BlockFile interface {
	NextBlockNumber() uint64

	Get(kind string, number uint64) ([]byte, error)

	AppendBlock(number uint64, hash, header, extra, receipts, transactions []byte) (err error)

	BatchAppendBlock(number uint64, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions [][]byte) (err error)

	TruncateBlocks(targetBlock uint64) error

	Close() error
}

type blockFile struct {
	nextBlockNumber uint64 // next block number

	tables       map[string]*BlockTable // Data tables for store nextBlockNumber
	instanceLock fileutil.Releaser      // File-system lock to prevent double opens

	logger    logrus.FieldLogger
	closeOnce sync.Once

	appendLock sync.Mutex
}

func NewBlockFile(p string, logger logrus.FieldLogger) (BlockFile, error) {
	return newBlockFile(p, logger)
}

func newBlockFile(p string, logger logrus.FieldLogger) (*blockFile, error) {
	if info, err := os.Lstat(p); !os.IsNotExist(err) {
		if info.Mode()&os.ModeSymlink != 0 {
			logger.WithField("path", p).Error("Symbolic link is not supported")
			return nil, errors.New("symbolic link datadir is not supported")
		}
	}
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return nil, err
	}
	lock, _, err := fileutil.Flock(filepath.Join(p, "FLOCK"))
	if err != nil {
		return nil, err
	}
	blockfile := &blockFile{
		tables:       make(map[string]*BlockTable),
		instanceLock: lock,
		logger:       logger,
	}
	for name := range BlockFileSchema {
		table, err := newTable(p, name, 2*1000*1000*1000, logger)
		if err != nil {
			for _, table := range blockfile.tables {
				_ = table.Close()
			}
			_ = lock.Release()
			return nil, err
		}
		blockfile.tables[name] = table
	}
	if err := blockfile.repair(); err != nil {
		for _, table := range blockfile.tables {
			_ = table.Close()
		}
		_ = lock.Release()
		return nil, err
	}

	return blockfile, nil
}

func (bf *blockFile) NextBlockNumber() uint64 {
	return atomic.LoadUint64(&bf.nextBlockNumber)
}

func (bf *blockFile) Get(kind string, number uint64) ([]byte, error) {
	if table := bf.tables[kind]; table != nil {
		return table.Retrieve(number)
	}
	return nil, errors.New("unknown table")
}

func (bf *blockFile) AppendBlock(number uint64, hash, header, extra, receipts, transactions []byte) (err error) {
	bf.appendLock.Lock()
	defer bf.appendLock.Unlock()

	err = bf.doBatchAppendBlock(number, [][]byte{header}, [][]byte{extra}, [][]byte{receipts}, [][]byte{transactions})
	if err != nil {
		bf.logger.WithFields(logrus.Fields{
			"number": bf.nextBlockNumber,
			"hash":   hash,
			"err":    err,
		}).Error("Failed to append block")
	}
	return err
}

func (bf *blockFile) BatchAppendBlock(number uint64, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions [][]byte) (err error) {
	bf.appendLock.Lock()
	defer bf.appendLock.Unlock()
	err = bf.doBatchAppendBlock(number, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions)
	if err != nil {
		bf.logger.WithFields(logrus.Fields{
			"number":              bf.nextBlockNumber,
			"first_hash":          listOfHash[0],
			"batch_append_number": len(listOfHash),
			"err":                 err,
		}).Error("Failed to batch append block")
	}
	return err
}

func (bf *blockFile) doBatchAppendBlock(number uint64, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions [][]byte) error {
	if atomic.LoadUint64(&bf.nextBlockNumber) != number {
		return errors.New("the append operation is out-order")
	}

	batchNum := len(listOfHeader)
	if batchNum == 0 {
		return errors.New("empty doBatch append block data")
	}
	if !(batchNum == len(listOfExtra) && batchNum == len(listOfReceipts) && batchNum == len(listOfTransactions)) {
		return errors.New("doBatch append block data param's length not match")
	}

	var err error
	defer func() {
		if err != nil {
			rerr := bf.repair()
			if rerr != nil {
				bf.logger.WithField("err", err).Errorf("Failed to repair blockfile")
			}
			bf.logger.WithFields(logrus.Fields{
				"number": number,
				"err":    err,
			}).Info("Append block failed")
		}
	}()
	if err = bf.tables[BlockFileHeaderTable].BatchAppend(bf.nextBlockNumber, listOfHeader); err != nil {
		return errors.Wrap(err, "failed to append block header")
	}
	if err = bf.tables[BlockFileTXsTable].BatchAppend(bf.nextBlockNumber, listOfTransactions); err != nil {
		return errors.Wrap(err, "failed to append block transactions")
	}
	if err = bf.tables[BlockFileExtraTable].BatchAppend(bf.nextBlockNumber, listOfExtra); err != nil {
		return errors.Wrap(err, "failed to append block extra")
	}
	if err = bf.tables[BlockFileReceiptsTable].BatchAppend(bf.nextBlockNumber, listOfReceipts); err != nil {
		return errors.Wrap(err, "failed to append block receipts")
	}
	atomic.AddUint64(&bf.nextBlockNumber, uint64(batchNum)) // Only modify atomically
	return nil
}

func (bf *blockFile) TruncateBlocks(targetBlock uint64) error {
	if targetBlock >= atomic.LoadUint64(&bf.nextBlockNumber) {
		return nil
	}
	for _, table := range bf.tables {
		if err := table.truncate(targetBlock + 1); err != nil {
			return err
		}
	}
	atomic.StoreUint64(&bf.nextBlockNumber, targetBlock+1)
	return nil
}

// repair truncates all data tables to the same length.
func (bf *blockFile) repair() error {
	minNumber := uint64(math.MaxUint64)
	for _, table := range bf.tables {
		items := atomic.LoadUint64(&table.items)
		if minNumber > items {
			minNumber = items
		}
	}
	for _, table := range bf.tables {
		if err := table.truncate(minNumber); err != nil {
			return err
		}
	}
	atomic.StoreUint64(&bf.nextBlockNumber, minNumber)
	return nil
}

func (bf *blockFile) Close() error {
	var errs []error
	bf.closeOnce.Do(func() {
		for _, table := range bf.tables {
			if err := table.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if err := bf.instanceLock.Release(); err != nil {
			errs = append(errs, err)
		}
	})
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
