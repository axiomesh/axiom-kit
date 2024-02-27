package blockfile

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	pkgerr "github.com/pkg/errors"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/sirupsen/logrus"
)

type BlockFile struct {
	blocks uint64 // Number of blocks

	tables       map[string]*BlockTable // Data tables for stroring blocks
	instanceLock fileutil.Releaser      // File-system lock to prevent double opens

	logger    logrus.FieldLogger
	closeOnce sync.Once

	appendLock sync.Mutex
}

func NewBlockFile(p string, logger logrus.FieldLogger) (*BlockFile, error) {
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
	blockfile := &BlockFile{
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

func (bf *BlockFile) Blocks() (uint64, error) {
	return atomic.LoadUint64(&bf.blocks), nil
}

func (bf *BlockFile) Get(kind string, number uint64) ([]byte, error) {
	if table := bf.tables[kind]; table != nil {
		return table.Retrieve(number - 1)
	}
	return nil, errors.New("unknown table")
}

func (bf *BlockFile) AppendBlock(number uint64, hash, body, receipts, transactions []byte) (err error) {
	bf.appendLock.Lock()
	defer bf.appendLock.Unlock()

	err = bf.doBatchAppendBlock(number, [][]byte{hash}, [][]byte{body}, [][]byte{receipts}, [][]byte{transactions})
	if err != nil {
		bf.logger.WithFields(logrus.Fields{
			"block_number": bf.blocks,
			"hash":         hash,
			"err":          err,
		}).Error("Failed to append block")
	}
	return err
}

func (bf *BlockFile) BatchAppendBlock(number uint64, listOfHash, listOfBody, listOfReceipts, listOfTransactions [][]byte) (err error) {
	bf.appendLock.Lock()
	defer bf.appendLock.Unlock()
	batchNum := len(listOfHash)
	if batchNum == 0 {
		return errors.New("empty batch append block data")
	}
	if !(batchNum == len(listOfBody) && batchNum == len(listOfReceipts) && batchNum == len(listOfTransactions)) {
		return errors.New("batch append block data param's length not match")
	}
	err = bf.doBatchAppendBlock(number, listOfHash, listOfBody, listOfReceipts, listOfTransactions)
	if err != nil {
		bf.logger.WithFields(logrus.Fields{
			"block_number":        bf.blocks,
			"first_hash":          listOfHash[0],
			"batch_append_number": batchNum,
			"err":                 err,
		}).Error("Failed to batch append block")
	}
	return err
}

func (bf *BlockFile) doBatchAppendBlock(number uint64, listOfHash, listOfBody, listOfReceipts, listOfTransactions [][]byte) error {
	if atomic.LoadUint64(&bf.blocks) != number {
		return errors.New("the append operation is out-order")
	}

	batchNum := len(listOfHash)
	if batchNum == 0 {
		return errors.New("empty doBatch append block data")
	}
	if !(batchNum == len(listOfBody) && batchNum == len(listOfReceipts) && batchNum == len(listOfTransactions)) {
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
	if err = bf.tables[BlockFileHashTable].BatchAppend(bf.blocks, listOfHash); err != nil {
		return pkgerr.Wrap(err, "failed to append block hash")
	}
	if err = bf.tables[BlockFileBodiesTable].BatchAppend(bf.blocks, listOfBody); err != nil {
		return pkgerr.Wrap(err, "failed to append block body")
	}
	if err = bf.tables[BlockFileTXsTable].BatchAppend(bf.blocks, listOfTransactions); err != nil {
		return pkgerr.Wrap(err, "failed to append block transactions")
	}
	if err = bf.tables[BlockFileReceiptTable].BatchAppend(bf.blocks, listOfReceipts); err != nil {
		return pkgerr.Wrap(err, "failed to append block receipt")
	}
	atomic.AddUint64(&bf.blocks, uint64(batchNum)) // Only modify atomically
	return nil
}

func (bf *BlockFile) TruncateBlocks(items uint64) error {
	if atomic.LoadUint64(&bf.blocks) <= items {
		return nil
	}
	for _, table := range bf.tables {
		if err := table.truncate(items); err != nil {
			return err
		}
	}
	atomic.StoreUint64(&bf.blocks, items)
	return nil
}

// repair truncates all data tables to the same length.
func (bf *BlockFile) repair() error {
	min := uint64(math.MaxUint64)
	for _, table := range bf.tables {
		items := atomic.LoadUint64(&table.items)
		if min > items {
			min = items
		}
	}
	for _, table := range bf.tables {
		if err := table.truncate(min); err != nil {
			return err
		}
	}
	atomic.StoreUint64(&bf.blocks, min)
	return nil
}

func (bf *BlockFile) Close() error {
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
