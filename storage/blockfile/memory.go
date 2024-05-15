package blockfile

import (
	"sync"

	"github.com/pkg/errors"
)

type memory struct {
	nextBlockNumber uint64
	tables          map[string]map[uint64][]byte
	lock            sync.RWMutex
}

func NewMemory() BlockFile {
	tables := make(map[string]map[uint64][]byte)
	for name := range BlockFileSchema {
		tables[name] = map[uint64][]byte{}
	}

	return &memory{
		nextBlockNumber: 0,
		tables:          tables,
		lock:            sync.RWMutex{},
	}
}

func (m *memory) NextBlockNumber() uint64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.nextBlockNumber
}

func (m *memory) Get(kind string, number uint64) ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if table := m.tables[kind]; table != nil {
		if number < m.nextBlockNumber {
			return table[number], nil
		}
		return nil, errors.New("out of bounds")
	}
	return nil, errors.New("unknown table")
}

func (m *memory) AppendBlock(number uint64, hash, header, extra, receipts, transactions []byte) (err error) {
	return m.BatchAppendBlock(number, [][]byte{hash}, [][]byte{header}, [][]byte{extra}, [][]byte{receipts}, [][]byte{transactions})
}

func (m *memory) BatchAppendBlock(number uint64, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions [][]byte) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.nextBlockNumber != number {
		return errors.New("the append operation is out-order")
	}

	batchNum := len(listOfHeader)
	if batchNum == 0 {
		return errors.New("empty doBatch append block data")
	}
	if !(batchNum == len(listOfExtra) && batchNum == len(listOfReceipts) && batchNum == len(listOfTransactions) && batchNum == len(listOfHash)) {
		return errors.New("doBatch append block data param's length not match")
	}

	for i := 0; i < batchNum; i++ {
		blockNumber := m.nextBlockNumber + uint64(i)
		m.tables[BlockFileHeaderTable][blockNumber] = listOfHeader[i]
		m.tables[BlockFileTXsTable][blockNumber] = listOfTransactions[i]
		m.tables[BlockFileExtraTable][blockNumber] = listOfExtra[i]
		m.tables[BlockFileReceiptsTable][blockNumber] = listOfReceipts[i]
	}
	m.nextBlockNumber += uint64(batchNum)
	return nil
}

func (m *memory) TruncateBlocks(targetBlock uint64) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if targetBlock >= m.nextBlockNumber {
		return nil
	}

	for i := m.nextBlockNumber - 1; i > targetBlock; i-- {
		delete(m.tables[BlockFileHeaderTable], i)
		delete(m.tables[BlockFileTXsTable], i)
		delete(m.tables[BlockFileExtraTable], i)
		delete(m.tables[BlockFileReceiptsTable], i)
	}
	m.nextBlockNumber = targetBlock + 1

	return nil
}

func (m *memory) Close() error {
	return nil
}
