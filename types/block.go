package types

import (
	"crypto/sha256"
	"github.com/samber/lo"
	"sync/atomic"

	"github.com/axiomesh/axiom-kit/types/pb"
)

type BlockHeader struct {
	Number          uint64
	StateRoot       *Hash
	TxRoot          *Hash
	ReceiptRoot     *Hash
	ParentHash      *Hash
	Timestamp       int64
	Epoch           uint64
	Bloom           *Bloom
	GasPrice        int64
	GasUsed         uint64
	ProposerAccount string
	ProposerNodeID  uint64
	Extra           []byte

	hashCache atomic.Value
}

func (h *BlockHeader) toPB() (*pb.BlockHeader, error) {
	if h == nil {
		return nil, nil
	}
	pbHeader := &pb.BlockHeader{
		Number:          h.Number,
		Timestamp:       h.Timestamp,
		Epoch:           h.Epoch,
		GasPrice:        h.GasPrice,
		ProposerAccount: h.ProposerAccount,
		GasUsed:         h.GasUsed,
		ProposerNodeId:  h.ProposerNodeID,
		Extra:           h.Extra,
	}
	if h.StateRoot != nil {
		pbHeader.StateRoot = h.StateRoot.Bytes()
	}
	if h.TxRoot != nil {
		pbHeader.TxRoot = h.TxRoot.Bytes()
	}
	if h.ReceiptRoot != nil {
		pbHeader.ReceiptRoot = h.ReceiptRoot.Bytes()
	}
	if h.ParentHash != nil {
		pbHeader.ParentHash = h.ParentHash.Bytes()
	}
	if h.Bloom != nil {
		pbHeader.Bloom = h.Bloom.Bytes()
	}
	return pbHeader, nil
}

func (h *BlockHeader) fromPB(m *pb.BlockHeader) error {
	var err error
	h.Number = m.Number
	if len(m.StateRoot) != 0 {
		h.StateRoot, err = decodeHash(m.StateRoot)
		if err != nil {
			return err
		}
	}
	if len(m.TxRoot) != 0 {
		h.TxRoot, err = decodeHash(m.TxRoot)
		if err != nil {
			return err
		}
	}
	if len(m.ReceiptRoot) != 0 {
		h.ReceiptRoot, err = decodeHash(m.ReceiptRoot)
		if err != nil {
			return err
		}
	}
	if len(m.ParentHash) != 0 {
		h.ParentHash, err = decodeHash(m.ParentHash)
		if err != nil {
			return err
		}
	}
	h.Timestamp = m.Timestamp
	h.Epoch = m.Epoch
	h.ProposerAccount = m.ProposerAccount
	if len(m.Bloom) != 0 {
		h.Bloom, err = decodeBloom(m.Bloom)
		if err != nil {
			return err
		}
	}
	h.GasPrice = m.GasPrice
	h.GasUsed = m.GasUsed
	h.ProposerNodeID = m.ProposerNodeId
	h.Extra = m.Extra
	return nil
}

func (h *BlockHeader) Marshal() ([]byte, error) {
	helper, err := h.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (h *BlockHeader) Unmarshal(data []byte) error {
	helper := pb.BlockHeaderFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return h.fromPB(helper)
}

func (h *BlockHeader) CalculateHash() *Hash {
	blockheader := &BlockHeader{
		Number:          h.Number,
		StateRoot:       h.StateRoot,
		TxRoot:          h.TxRoot,
		ReceiptRoot:     h.ReceiptRoot,
		ParentHash:      h.ParentHash,
		Timestamp:       h.Timestamp,
		Epoch:           h.Epoch,
		GasPrice:        h.GasPrice,
		GasUsed:         h.GasUsed,
		ProposerAccount: h.ProposerAccount,
		ProposerNodeID:  h.ProposerNodeID,
		Extra:           h.Extra,
		hashCache:       atomic.Value{},
	}
	raw, err := blockheader.Marshal()
	if err != nil {
		panic(err)
	}

	data := sha256.Sum256(raw)

	return NewHash(data[:])
}

func (h *BlockHeader) Hash() *Hash {
	if h == nil {
		return nil
	}

	if hash := h.hashCache.Load(); hash != nil {
		return hash.(*Hash)
	}

	res := h.CalculateHash()
	h.hashCache.Store(res)
	return res
}

func (h *BlockHeader) Clone() *BlockHeader {
	if h == nil {
		return nil
	}
	bl := &Bloom{}
	if h.Bloom != nil {
		bl.SetBytes(h.Bloom.Bytes())
	}

	return &BlockHeader{
		Number:          h.Number,
		StateRoot:       h.StateRoot.Clone(),
		TxRoot:          h.TxRoot.Clone(),
		ReceiptRoot:     h.ReceiptRoot.Clone(),
		ParentHash:      h.ParentHash.Clone(),
		Timestamp:       h.Timestamp,
		Epoch:           h.Epoch,
		Bloom:           bl,
		GasPrice:        h.GasPrice,
		GasUsed:         h.GasUsed,
		ProposerAccount: h.ProposerAccount,
		ProposerNodeID:  h.ProposerNodeID,
		Extra:           h.Extra,
		hashCache:       atomic.Value{},
	}
}

// For TIMC and system contract
type BlockExtra struct {
}

func (b *BlockExtra) toPB() (*pb.BlockExtra, error) {
	if b == nil {
		return nil, nil
	}
	return &pb.BlockExtra{}, nil
}

func (b *BlockExtra) fromPB(m *pb.BlockExtra) error {
	return nil
}

func (b *BlockExtra) Marshal() ([]byte, error) {
	helper, err := b.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (b *BlockExtra) Unmarshal(data []byte) error {
	helper := pb.BlockExtraFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return b.fromPB(helper)
}

func (b *BlockExtra) Clone() *BlockExtra {
	if b == nil {
		return nil
	}
	return &BlockExtra{}
}

type Block struct {
	Header       *BlockHeader
	Transactions []*Transaction
	Extra        *BlockExtra
}

func (b *Block) toPB() (*pb.Block, error) {
	if b == nil {
		return nil, nil
	}

	headerPB, err := b.Header.toPB()
	if err != nil {
		return nil, err
	}
	extraPB, err := b.Extra.toPB()
	if err != nil {
		return nil, err
	}
	var txsRaw [][]byte
	for _, tx := range b.Transactions {
		txRaw, err := tx.Marshal()
		if err != nil {
			return nil, err
		}
		txsRaw = append(txsRaw, txRaw)
	}
	return &pb.Block{
		Header:       headerPB,
		Transactions: txsRaw,
		Extra:        extraPB,
	}, nil
}

func (b *Block) fromPB(m *pb.Block) error {
	var err error
	if m.Header != nil {
		b.Header = &BlockHeader{}
		if err = b.Header.fromPB(m.Header); err != nil {
			return err
		}
	}
	if m.Extra != nil {
		b.Extra = &BlockExtra{}
		if err = b.Extra.fromPB(m.Extra); err != nil {
			return err
		}
	}
	for _, txRaw := range m.Transactions {
		tx := &Transaction{}
		if err := tx.Unmarshal(txRaw); err != nil {
			return err
		}
		b.Transactions = append(b.Transactions, tx)
	}
	return nil
}

func (b *Block) Marshal() ([]byte, error) {
	helper, err := b.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (b *Block) Unmarshal(data []byte) error {
	helper := pb.BlockFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return b.fromPB(helper)
}

func (b *Block) Hash() *Hash {
	if b == nil {
		return nil
	}

	return b.Header.Hash()
}

func (b *Block) Height() uint64 {
	if b == nil || b.Header == nil {
		return 0
	}

	return b.Header.Number
}

func (b *Block) Size() int {
	helper, err := b.toPB()
	if err != nil {
		return 0
	}
	return helper.SizeVT()
}

func (b *Block) Clone() *Block {
	if b == nil {
		return nil
	}

	txs := make([]*Transaction, len(b.Transactions))
	lo.ForEach(b.Transactions, func(tx *Transaction, i int) {
		txs[i] = tx.Clone()
	})

	return &Block{
		Header:       b.Header.Clone(),
		Transactions: txs,
		Extra:        b.Extra.Clone(),
	}
}

type BlockBody struct {
	Transactions []*Transaction
	Extra        *BlockExtra
}

func (b *BlockBody) toPB() (*pb.BlockBody, error) {
	if b == nil {
		return nil, nil
	}
	extraPB, err := b.Extra.toPB()
	if err != nil {
		return nil, err
	}
	var txsRaw [][]byte
	for _, tx := range b.Transactions {
		txRaw, err := tx.Marshal()
		if err != nil {
			return nil, err
		}
		txsRaw = append(txsRaw, txRaw)
	}
	return &pb.BlockBody{
		Transactions: txsRaw,
		Extra:        extraPB,
	}, nil
}

func (b *BlockBody) fromPB(m *pb.BlockBody) error {
	var err error
	if m.Extra != nil {
		b.Extra = &BlockExtra{}
		if err = b.Extra.fromPB(m.Extra); err != nil {
			return err
		}
	}
	for _, txRaw := range m.Transactions {
		tx := &Transaction{}
		if err := tx.Unmarshal(txRaw); err != nil {
			return err
		}
		b.Transactions = append(b.Transactions, tx)
	}
	return nil
}

func (b *BlockBody) Marshal() ([]byte, error) {
	helper, err := b.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (b *BlockBody) Unmarshal(data []byte) error {
	helper := pb.BlockBodyFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return b.fromPB(helper)
}

func (b *BlockBody) Clone() *BlockBody {
	if b == nil {
		return nil
	}
	txs := make([]*Transaction, len(b.Transactions))
	lo.ForEach(b.Transactions, func(tx *Transaction, i int) {
		txs[i] = tx.Clone()
	})
	return &BlockBody{
		Transactions: txs,
		Extra:        b.Extra.Clone(),
	}
}
