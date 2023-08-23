package types

import (
	"math/big"

	"github.com/axiomesh/axiom-kit/types/pb"
)

type ChainMeta struct {
	Height uint64
	// GasPrice is the next block's price
	// is different from gas price in the block header
	// which means the next block's gas price
	GasPrice  *big.Int
	BlockHash *Hash
}

func (m *ChainMeta) toPB() (*pb.ChainMeta, error) {
	if m == nil {
		return &pb.ChainMeta{}, nil
	}
	return &pb.ChainMeta{
		Height:    m.Height,
		GasPrice:  m.GasPrice.Bytes(),
		BlockHash: m.BlockHash.Bytes(),
	}, nil
}

func (m *ChainMeta) fromPB(p *pb.ChainMeta) error {
	var err error
	m.BlockHash, err = decodeHash(p.BlockHash)
	if err != nil {
		return err
	}
	m.Height = p.Height
	m.GasPrice = new(big.Int).SetBytes(p.GasPrice)
	return nil
}

func (m *ChainMeta) Marshal() ([]byte, error) {
	helper, err := m.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (m *ChainMeta) Unmarshal(data []byte) error {
	helper := &pb.ChainMeta{}
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return m.fromPB(helper)
}

type VpInfo struct {
	Id      uint64
	Pid     string
	Account string
	Hosts   []string
}

func (i *VpInfo) toPB() (*pb.VpInfo, error) {
	if i == nil {
		return &pb.VpInfo{}, nil
	}
	return &pb.VpInfo{
		Id:      i.Id,
		Pid:     i.Pid,
		Account: i.Account,
		Hosts:   i.Hosts,
	}, nil
}

func (i *VpInfo) fromPB(m *pb.VpInfo) error {
	i.Id = m.Id
	i.Pid = m.Pid
	i.Account = m.Account
	i.Hosts = m.Hosts
	return nil
}

func (i *VpInfo) Marshal() ([]byte, error) {
	helper, err := i.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (i *VpInfo) Unmarshal(data []byte) error {
	helper := &pb.VpInfo{}
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}

	return i.fromPB(helper)
}

type TransactionMeta struct {
	BlockHash   *Hash
	BlockHeight uint64
	Index       uint64
}

func (m *TransactionMeta) toPB() (*pb.TransactionMeta, error) {
	if m == nil {
		return &pb.TransactionMeta{}, nil
	}

	return &pb.TransactionMeta{
		BlockHash:   m.BlockHash.Bytes(),
		BlockHeight: m.BlockHeight,
		Index:       m.Index,
	}, nil
}

func (m *TransactionMeta) fromPB(p *pb.TransactionMeta) error {
	var err error
	m.BlockHash, err = decodeHash(p.BlockHash)
	if err != nil {
		return err
	}
	m.BlockHeight = p.BlockHeight
	m.Index = p.Index
	return nil
}

func (m *TransactionMeta) Marshal() ([]byte, error) {
	helper, err := m.toPB()
	if err != nil {
		return nil, err
	}
	return helper.MarshalVTStrict()
}

func (m *TransactionMeta) Unmarshal(data []byte) error {
	helper := &pb.TransactionMeta{}
	err := helper.UnmarshalVT(data)
	if err != nil {
		return err
	}
	return m.fromPB(helper)
}
