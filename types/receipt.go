package types

import (
	"crypto/sha256"
	"encoding/json"
	"math/big"

	"github.com/samber/lo"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types/pb"
)

type ReceiptStatus int32

const (
	ReceiptSUCCESS ReceiptStatus = 0
	ReceiptFAILED  ReceiptStatus = 1
)

var receiptStatusName = map[ReceiptStatus]string{
	ReceiptSUCCESS: "SUCCESS",
	ReceiptFAILED:  "FAILED",
}

func (x ReceiptStatus) String() string {
	return receiptStatusName[x]
}

func (x ReceiptStatus) toPB() pb.Receipt_Status {
	switch x {
	case ReceiptSUCCESS:
		return pb.Receipt_SUCCESS
	case ReceiptFAILED:
		return pb.Receipt_FAILED
	default:
		return pb.Receipt_FAILED
	}
}

func receiptStatusFromPB(x pb.Receipt_Status) ReceiptStatus {
	switch x {
	case pb.Receipt_SUCCESS:
		return ReceiptSUCCESS
	case pb.Receipt_FAILED:
		return ReceiptFAILED
	default:
		return ReceiptFAILED
	}
}

type EvmLog struct {
	Address          *Address
	Topics           []*Hash
	Data             []byte
	BlockNumber      uint64
	TransactionHash  *Hash
	TransactionIndex uint64
	BlockHash        *Hash
	LogIndex         uint64
	Removed          bool
}

func (l *EvmLog) toPB() (*pb.EvmLog, error) {
	if l == nil {
		return &pb.EvmLog{}, nil
	}

	return &pb.EvmLog{
		Address: l.Address.Bytes(),
		Topics: lo.Map(l.Topics, func(item *Hash, index int) []byte {
			return item.Bytes()
		}),
		Data:             l.Data,
		BlockNumber:      l.BlockNumber,
		TransactionHash:  l.TransactionHash.Bytes(),
		TransactionIndex: l.TransactionIndex,
		BlockHash:        l.BlockHash.Bytes(),
		LogIndex:         l.LogIndex,
		Removed:          l.Removed,
	}, nil
}

func (l *EvmLog) fromPB(p *pb.EvmLog) error {
	var err error
	l.Address, err = decodeAddress(p.Address)
	if err != nil {
		return err
	}
	for _, pTopic := range p.Topics {
		topic, err := decodeHash(pTopic)
		if err != nil {
			return err
		}
		l.Topics = append(l.Topics, topic)
	}
	l.Data = p.Data
	l.BlockNumber = p.BlockNumber
	l.TransactionHash, err = decodeHash(p.TransactionHash)
	if err != nil {
		return err
	}
	l.TransactionIndex = p.TransactionIndex
	l.BlockHash, err = decodeHash(p.BlockHash)
	if err != nil {
		return err
	}
	l.LogIndex = p.LogIndex
	l.Removed = p.Removed
	return nil
}

func (l *EvmLog) Marshal() ([]byte, error) {
	helper, err := l.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (l *EvmLog) Unmarshal(data []byte) error {
	helper := pb.EvmLogFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}
	return l.fromPB(helper)
}

type logEncoder struct {
	Address     Address
	Topics      []*Hash
	Data        hexutil.Bytes
	BlockNumber hexutil.Uint64
	TxHash      Hash
	TxIndex     hexutil.Uint
	BlockHash   Hash
	Index       hexutil.Uint
	Removed     bool
}

// MarshalJSON marshals as JSON.
func (l *EvmLog) MarshalJSON() ([]byte, error) {
	var enc logEncoder

	if l.Address != nil {
		enc.Address = *l.Address
	}
	enc.Topics = l.Topics
	enc.Data = l.Data
	enc.BlockNumber = hexutil.Uint64(l.BlockNumber)
	if l.TransactionHash != nil {
		enc.TxHash = *l.TransactionHash
	}
	enc.TxIndex = hexutil.Uint(l.TransactionIndex)
	if l.BlockHash != nil {
		enc.BlockHash = *l.BlockHash
	}
	enc.Index = hexutil.Uint(l.LogIndex)
	enc.Removed = l.Removed

	return json.Marshal(&enc)
}

type Receipt struct {
	TxHash            *Hash
	Ret               []byte
	Status            ReceiptStatus
	GasUsed           uint64
	CumulativeGasUsed uint64
	EffectiveGasPrice *big.Int
	EvmLogs           []*EvmLog
	Bloom             *Bloom
	ContractAddress   *Address
}

func (r *Receipt) toPB() (*pb.Receipt, error) {
	if r == nil {
		return &pb.Receipt{}, nil
	}

	evmLogs := make([]*pb.EvmLog, len(r.EvmLogs))
	for i, l := range r.EvmLogs {
		log, err := l.toPB()
		if err != nil {
			return nil, err
		}
		evmLogs[i] = log
	}
	return &pb.Receipt{
		TxHash:            r.TxHash.Bytes(),
		Ret:               r.Ret,
		Status:            r.Status.toPB(),
		GasUsed:           r.GasUsed,
		CumulativeGasUsed: r.CumulativeGasUsed,
		EffectiveGasPrice: r.EffectiveGasPrice.Bytes(),
		EvmLogs:           evmLogs,
		Bloom:             r.Bloom.Bytes(),
		ContractAddress:   r.ContractAddress.Bytes(),
	}, nil
}

func (r *Receipt) fromPB(p *pb.Receipt) error {
	var err error
	r.TxHash, err = decodeHash(p.TxHash)
	if err != nil {
		return err
	}
	r.Ret = p.Ret
	r.Status = receiptStatusFromPB(p.Status)
	r.GasUsed = p.GasUsed
	for _, l := range p.EvmLogs {
		log := &EvmLog{}
		if err := log.fromPB(l); err != nil {
			return err
		}
		r.EvmLogs = append(r.EvmLogs, log)
	}
	r.Bloom, err = decodeBloom(p.Bloom)
	if err != nil {
		return err
	}
	r.ContractAddress, err = decodeAddress(p.ContractAddress)
	if err != nil {
		return err
	}
	r.CumulativeGasUsed = p.CumulativeGasUsed
	if len(p.EffectiveGasPrice) != 0 {
		r.EffectiveGasPrice = big.NewInt(0).SetBytes(p.EffectiveGasPrice)
	}
	return nil
}

func (r *Receipt) Marshal() ([]byte, error) {
	helper, err := r.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (r *Receipt) Unmarshal(data []byte) error {
	helper := pb.ReceiptFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}
	return r.fromPB(helper)
}

func (r *Receipt) Hash() *Hash {
	receipt := &Receipt{
		TxHash:            r.TxHash,
		Ret:               r.Ret,
		Status:            r.Status,
		EvmLogs:           r.EvmLogs,
		Bloom:             r.Bloom,
		GasUsed:           r.GasUsed,
		CumulativeGasUsed: r.CumulativeGasUsed,
		EffectiveGasPrice: r.EffectiveGasPrice,
		ContractAddress:   r.ContractAddress,
	}
	body, err := receipt.Marshal()
	if err != nil {
		panic(err)
	}

	data := sha256.Sum256(body)

	return NewHash(data[:])
}

func (r *Receipt) IsSuccess() bool {
	return r.Status == ReceiptSUCCESS
}
