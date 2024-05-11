package mock_txpool

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/axiomesh/axiom-kit/txpool"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/axiomesh/axiom-kit/types"
)

type MockMinimalTxPool[T any, Constraint types.TXConstraint[T]] struct {
	*MockTxPool[T, Constraint]
	allTxs                map[string]map[string]*T // account -> txHash -> tx
	batchCache            map[string]*txpool.RequestHashBatch[T, Constraint]
	missingBatchCache     map[string]map[uint64]string
	batchedTxs            map[string]bool
	batchSize             int
	noBatchSize           *atomic.Int64
	notifyGenerateBatch   bool
	notifyGenerateBatchFn func(typ int)
	notifyFindNextBatchFn func(completionMissingBatchHashes ...string) // notify consensus that it can find next batch
	started               atomic.Bool
}

// NewMockMinimalTxPool returns a minimal implement of MockTxPool which accepts
// any kinds of input and returns 'zero value' as all outputs.
// Users can define custom MockTxPool like this:
// func NewMockCustomTxPool(ctrl *gomock.Controller) *MockTxPool {...}
// in which users must specify output for all functions.
func NewMockMinimalTxPool[T any, Constraint types.TXConstraint[T]](batchSize int, ctrl *gomock.Controller) txpool.TxPool[T, Constraint] {
	mock := &MockMinimalTxPool[T, Constraint]{
		MockTxPool:        NewMockTxPool[T, Constraint](ctrl),
		allTxs:            make(map[string]map[string]*T),
		batchCache:        make(map[string]*txpool.RequestHashBatch[T, Constraint]),
		missingBatchCache: make(map[string]map[uint64]string),
		batchedTxs:        make(map[string]bool),
		batchSize:         batchSize,
		noBatchSize:       &atomic.Int64{},
	}
	mock.EXPECT().GenerateRequestBatch(gomock.Any()).DoAndReturn(func(typ int) (*txpool.RequestHashBatch[T, Constraint], error) {
		txList := make([]*T, 0)
		txHashList := make([]string, 0)
		localList := make([]bool, 0)
		switch typ {
		case txpool.GenBatchSizeEvent, txpool.GenBatchFirstEvent:
			if int(mock.noBatchSize.Load()) < mock.batchSize {
				return nil, fmt.Errorf("actual batch size %d is smaller than %d, ignore generate batch",
					mock.noBatchSize.Load(), mock.batchSize)
			}
		case txpool.GenBatchTimeoutEvent:
			if mock.noBatchSize.Load() == 0 {
				return nil, errors.New("there is no pending tx, ignore generate batch")
			}

		case txpool.GenBatchNoTxTimeoutEvent:
			if mock.noBatchSize.Load() != 0 {
				return nil, errors.New("there is pending tx, ignore generate empty batch")
			}
			newBatch := &txpool.RequestHashBatch[T, Constraint]{
				TxList:     make([]*T, 0),
				TxHashList: make([]string, 0),
				LocalList:  make([]bool, 0),
				Timestamp:  time.Now().UnixNano(),
			}
			batchHash := newBatch.GenerateBatchHash()
			newBatch.BatchHash = batchHash
			mock.batchCache[batchHash] = newBatch
			return newBatch, nil
		}
		for _, m := range mock.allTxs {
			for _, tx := range m {
				txHash := Constraint(tx).RbftGetTxHash()
				if mock.batchedTxs[txHash] {
					continue
				}
				txList = append(txList, tx)
				txHashList = append(txHashList, Constraint(tx).RbftGetTxHash())
				localList = append(localList, true)
				mock.batchedTxs[Constraint(tx).RbftGetTxHash()] = true
			}
		}
		if len(txList) == 0 {
			return nil, errors.New("there is no pending tx, ignore generate batch")
		}
		newBatch := &txpool.RequestHashBatch[T, Constraint]{
			TxList:     txList,
			TxHashList: txHashList,
			LocalList:  localList,
			Timestamp:  time.Now().UnixNano(),
		}
		mock.noBatchSize.Add(int64(-len(txList)))
		batchHash := newBatch.GenerateBatchHash()
		newBatch.BatchHash = batchHash
		mock.batchCache[batchHash] = newBatch

		if typ == txpool.GenBatchSizeEvent {
			mock.notifyGenerateBatch = false
		}
		return newBatch, nil
	}).AnyTimes()
	mock.EXPECT().AddLocalTx(gomock.Any()).DoAndReturn(func(tx *T) error {
		mock.addTx(tx)
		return nil
	}).AnyTimes()
	mock.EXPECT().AddRemoteTxs(gomock.Any()).Do(func(txs []*T) {
		lo.ForEach(txs, func(tx *T, _ int) {
			mock.addTx(tx)
		})
	}).AnyTimes()
	mock.EXPECT().AddRebroadcastTxs(gomock.Any()).Do(func(txs []*T) {
		lo.ForEach(txs, func(tx *T, _ int) {
			mock.addTx(tx)
		})
	}).AnyTimes()

	mock.EXPECT().GetRequestsByHashList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(batchHash string, timestamp int64, txHashList []string, _ []string) ([]*T, []bool, map[uint64]string, error) {
			if mock.batchCache[batchHash] != nil {
				return mock.batchCache[batchHash].TxList, mock.batchCache[batchHash].LocalList, nil, nil
			}
			missingTxHashList := make(map[uint64]string)
			txList := make([]*T, 0)
			localList := make([]bool, 0)
			lo.ForEach(txHashList, func(txHash string, index int) {
				exist := false
				for _, m := range mock.allTxs {
					if tx, ok := m[txHash]; ok {
						txList = append(txList, tx)
						localList = append(localList, true)
						mock.batchedTxs[txHash] = true
						exist = true
						break
					}
				}
				if !exist {
					missingTxHashList[uint64(index)] = txHash
				}
			})
			if len(missingTxHashList) != 0 {
				mock.missingBatchCache[batchHash] = missingTxHashList
				return nil, nil, missingTxHashList, nil
			}

			newBatch := &txpool.RequestHashBatch[T, Constraint]{
				BatchHash:  batchHash,
				TxList:     txList,
				TxHashList: txHashList,
				LocalList:  localList,
				Timestamp:  timestamp,
			}
			mock.noBatchSize.Add(int64(-len(txList)))
			mock.batchCache[batchHash] = newBatch
			return txList, localList, nil, nil
		}).AnyTimes()

	mock.EXPECT().RemoveBatches(gomock.Any()).Do(func(digests []string) {
		lo.ForEach(digests, func(digest string, _ int) {
			batch, ok := mock.batchCache[digest]
			if ok {
				lo.ForEach(batch.TxList, func(tx *T, _ int) {
					delete(mock.batchedTxs, Constraint(tx).RbftGetTxHash())
					if mock.allTxs != nil {
						delete(mock.allTxs[Constraint(tx).RbftGetFrom()], Constraint(tx).RbftGetTxHash())
						if len(mock.allTxs[Constraint(tx).RbftGetFrom()]) == 0 {
							delete(mock.allTxs, Constraint(tx).RbftGetFrom())
						}
					}
				})
			}
		})
	}).AnyTimes()
	mock.EXPECT().IsPoolFull().Return(false).AnyTimes()
	mock.EXPECT().HasPendingRequestInPool().DoAndReturn(func() bool {
		return len(mock.allTxs) > 0
	}).AnyTimes()
	mock.EXPECT().RestoreOneBatch(gomock.Any()).Return(nil).AnyTimes()

	mock.EXPECT().PendingRequestsNumberIsReady().Return(false).AnyTimes()

	mock.EXPECT().SendMissingRequests(gomock.Any(), gomock.Any()).DoAndReturn(func(batchHash string, missingHashList map[uint64]string) (map[uint64]*T, error) {
		batch := mock.batchCache[batchHash]
		missingTxList := make(map[uint64]*T)
		if batch == nil {
			return nil, errors.New("batch not found")
		}
		for index, hash := range missingHashList {
			tx := batch.TxList[index]
			if Constraint(tx).RbftGetTxHash() != hash {
				return nil, errors.New("hash not match")
			}
			missingTxList[index] = tx
		}
		return missingTxList, nil
	}).AnyTimes()
	mock.EXPECT().ReceiveMissingRequests(gomock.Any(), gomock.Any()).DoAndReturn(func(batchHash string, missingTxs map[uint64]*T) error {
		if missingHashList := mock.missingBatchCache[batchHash]; missingHashList != nil {
			for index, hash := range missingHashList {
				tx, ok := missingTxs[index]
				if !ok {
					return errors.New("missing tx not found")
				}
				if Constraint(tx).RbftGetTxHash() != hash {
					return errors.New("hash not match")
				}
				mock.addTx(tx)
			}
		} else {
			return errors.New("missing batch not found")
		}
		delete(mock.missingBatchCache, batchHash)
		return nil
	}).AnyTimes()
	mock.EXPECT().FilterOutOfDateRequests(gomock.Any()).DoAndReturn(func(_ bool) []*T {
		txList := make([]*T, 0)
		for _, m := range mock.allTxs {
			for _, tx := range m {
				txList = append(txList, tx)
			}
		}
		return txList
	}).AnyTimes()
	mock.EXPECT().RestorePool().Do(func() {
		for _, batch := range mock.batchCache {
			// move all batched txs to no batched txs
			lo.ForEach(batch.TxHashList, func(txHash string, index int) {
				mock.batchedTxs[txHash] = false
			})
			mock.noBatchSize.Add(int64(len(batch.TxHashList)))
		}
	}).AnyTimes()
	mock.EXPECT().ReConstructBatchByOrder(gomock.Any()).Return(nil, nil).AnyTimes()
	mock.EXPECT().Stop().AnyTimes()
	mock.EXPECT().Start().DoAndReturn(func() error {
		if mock.started.Load() {
			return errors.New("txpool already started")
		}
		mock.started.Store(true)
		return nil
	}).AnyTimes()
	mock.EXPECT().Init(gomock.Any()).Do(func(config txpool.ConsensusConfig) {
		mock.notifyGenerateBatchFn = config.NotifyGenerateBatchFn
		mock.notifyFindNextBatchFn = config.NotifyFindNextBatchFn
	}).AnyTimes()
	mock.EXPECT().GetLocalTxs().DoAndReturn(func() [][]byte {
		res := make([][]byte, 0)
		for _, txs := range mock.allTxs {
			for _, tx := range txs {
				marshal, _ := Constraint(tx).RbftMarshal()
				res = append(res, marshal)
			}
		}
		return res
	}).AnyTimes()
	mock.EXPECT().GetPendingTxByHash(gomock.Any()).DoAndReturn(func(hash string) *T {
		for _, txs := range mock.allTxs {
			for _, tx := range txs {
				if Constraint(tx).RbftGetTxHash() == hash {
					return tx
				}
			}
		}
		return nil
	}).AnyTimes()
	mock.EXPECT().IsStarted().DoAndReturn(func() bool {
		return mock.started.Load()
	}).AnyTimes()

	mock.EXPECT().ReplyBatchSignal().Return().AnyTimes()
	return mock
}

func (m *MockMinimalTxPool[T, Constraint]) addTx(tx *T) {
	account := Constraint(tx).RbftGetFrom()
	txHash := Constraint(tx).RbftGetTxHash()
	if m.allTxs[account] == nil {
		m.allTxs[account] = make(map[string]*T)
	}
	m.allTxs[account][txHash] = tx
	m.noBatchSize.Add(1)

	if !m.notifyGenerateBatch && int(m.noBatchSize.Load()) >= m.batchSize {
		m.notifyGenerateBatchFn(txpool.GenBatchSizeEvent)
		m.notifyGenerateBatch = true
	}
}

func (m *MockMinimalTxPool[T, Constraint]) SetBatchSize(size int) {
	m.batchSize = size
}
