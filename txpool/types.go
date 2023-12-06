package txpool

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"

	"github.com/axiomesh/axiom-kit/types"
)

type TxInfo[T any, Constraint types.TXConstraint[T]] struct {
	Tx          *T
	Local       bool
	LifeTime    int64
	ArrivedTime int64
}

type AccountMeta[T any, Constraint types.TXConstraint[T]] struct {
	CommitNonce  uint64
	PendingNonce uint64
	TxCount      uint64
	Txs          []*TxInfo[T, Constraint]
	SimpleTxs    []*TxSimpleInfo
}

type TxSimpleInfo struct {
	Hash        string
	Nonce       uint64
	Size        int
	Local       bool
	LifeTime    int64
	ArrivedTime int64
}

type BatchSimpleInfo struct {
	TxCount   uint64
	Txs       []*TxSimpleInfo
	Timestamp int64
}

type Meta[T any, Constraint types.TXConstraint[T]] struct {
	TxCountLimit    uint64
	TxCount         uint64
	ReadyTxCount    uint64
	Batches         map[string]*BatchSimpleInfo
	MissingBatchTxs map[string]map[uint64]string
	Accounts        map[string]*AccountMeta[T, Constraint]
}

// RequestHashBatch contains transactions that batched by primary.
type RequestHashBatch[T any, Constraint types.TXConstraint[T]] struct {
	BatchHash  string   // hash of this batch calculated by MD5
	TxHashList []string // list of all txs' hashes
	TxList     []*T     // list of all txs
	LocalList  []bool   // list track if tx is received locally or not
	Timestamp  int64    // generation time of this batch
}

type ConsensusConfig struct {
	SelfID                uint64
	NotifyGenerateBatchFn func(typ int)                                // notify consensus that it can generate a new batch
	NotifyFindNextBatchFn func(completionMissingBatchHashes ...string) // notify consensus that it can find next batch
}

type WrapperTxPointer struct {
	TxHash  string
	Account string
	Nonce   uint64
}

const (
	GenBatchTimeoutEvent = iota
	GenBatchNoTxTimeoutEvent
	GenBatchFirstEvent
	GenBatchSizeEvent
	ReConstructBatchEvent
	GetTxsForGenBatchEvent
)

// GetBatchHash calculate hash of a RequestHashBatch
func GetBatchHash[T any, Constraint types.TXConstraint[T]](batch *RequestHashBatch[T, Constraint]) string {
	h := md5.New()
	for _, hash := range batch.TxHashList {
		_, _ = h.Write([]byte(hash))
	}
	if batch.Timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(batch.Timestamp))
		_, _ = h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}