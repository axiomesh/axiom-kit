package types

import (
	"crypto/sha256"

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
}

func (h *BlockHeader) toPB() (*pb.BlockHeader, error) {
	if h == nil {
		return &pb.BlockHeader{}, nil
	}
	return &pb.BlockHeader{
		Number:          h.Number,
		StateRoot:       h.StateRoot.Bytes(),
		TxRoot:          h.TxRoot.Bytes(),
		ReceiptRoot:     h.ReceiptRoot.Bytes(),
		ParentHash:      h.ParentHash.Bytes(),
		Timestamp:       h.Timestamp,
		Epoch:           h.Epoch,
		Bloom:           h.Bloom.Bytes(),
		GasPrice:        h.GasPrice,
		GasUsed:         h.GasUsed,
		ProposerAccount: h.ProposerAccount,
	}, nil
}

func (h *BlockHeader) fromPB(m *pb.BlockHeader) error {
	var err error
	h.Number = m.Number
	h.StateRoot, err = decodeHash(m.StateRoot)
	if err != nil {
		return err
	}
	h.TxRoot, err = decodeHash(m.TxRoot)
	if err != nil {
		return err
	}
	h.ReceiptRoot, err = decodeHash(m.ReceiptRoot)
	if err != nil {
		return err
	}
	h.ParentHash, err = decodeHash(m.ParentHash)
	if err != nil {
		return err
	}
	h.Timestamp = m.Timestamp
	h.Epoch = m.Epoch
	h.ProposerAccount = m.ProposerAccount
	h.Bloom, err = decodeBloom(m.Bloom)
	if err != nil {
		return err
	}
	h.GasPrice = m.GasPrice
	h.GasUsed = m.GasUsed
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
	defer helper.ReturnToVTPool()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return h.fromPB(helper)
}

func (h *BlockHeader) Hash() *Hash {
	blockheader := &BlockHeader{
		Number:          h.Number,
		ParentHash:      h.ParentHash,
		StateRoot:       h.StateRoot,
		TxRoot:          h.TxRoot,
		ReceiptRoot:     h.ReceiptRoot,
		Epoch:           h.Epoch,
		GasPrice:        h.GasPrice,
		GasUsed:         h.GasUsed,
		ProposerAccount: h.ProposerAccount,
	}
	body, err := blockheader.Marshal()
	if err != nil {
		panic(err)
	}

	data := sha256.Sum256(body)

	return NewHash(data[:])
}

type Block struct {
	BlockHeader  *BlockHeader
	Transactions []*Transaction
	BlockHash    *Hash
	Signature    []byte
	Extra        []byte
}

func (b *Block) toPB() (*pb.Block, error) {
	if b == nil {
		return &pb.Block{}, nil
	}

	headerPB, err := b.BlockHeader.toPB()
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
		BlockHeader:  headerPB,
		Transactions: txsRaw,
		BlockHash:    b.BlockHash.Bytes(),
		Signature:    b.Signature,
		Extra:        b.Extra,
	}, nil
}

func (b *Block) fromPB(m *pb.Block) error {
	var err error
	b.BlockHeader = &BlockHeader{}
	if err = b.BlockHeader.fromPB(m.BlockHeader); err != nil {
		return err
	}
	for _, txRaw := range m.Transactions {
		tx := &Transaction{}
		if err := tx.Unmarshal(txRaw); err != nil {
			return err
		}
		b.Transactions = append(b.Transactions, tx)
	}
	b.BlockHash, err = decodeHash(m.BlockHash)
	if err != nil {
		return err
	}
	b.Signature = m.Signature
	b.Extra = m.Extra
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
	defer helper.ReturnToVTPool()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return b.fromPB(helper)
}

func (b *Block) Hash() *Hash {
	return b.BlockHeader.Hash()
}

func (b *Block) Height() uint64 {
	if b == nil || b.BlockHeader == nil {
		return 0
	}

	return b.BlockHeader.Number
}

func (b *Block) Size() int {
	helper, err := b.toPB()
	if err != nil {
		return 0
	}
	return helper.SizeVT()
}
