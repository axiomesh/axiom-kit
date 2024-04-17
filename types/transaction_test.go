package types

import (
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types/pb"
)

func TestEthTransaction_GetSignHash(t *testing.T) {
	rawTx := "0xf86c8085147d35700082520894f927bb571eaab8c9a361ab405c9e4891c5024380880de0b6b3a76400008025a00b8e3b66c1e7ae870802e3ef75f1ec741f19501774bd5083920ce181c2140b99a0040c122b7ebfb3d33813927246cbbad1c6bf210474f5d28053990abff0fd4f53"
	tx := &Transaction{}
	err := tx.Unmarshal(hexutil.Decode(rawTx))
	assert.Nil(t, err)
	addr := "0xC63573cB77ec56e0A1cb40199bb85838D71e4dce"

	InitEIP155Signer(big.NewInt(1))

	err = tx.VerifySignature()
	assert.Nil(t, err)

	sender, err := tx.sender()
	assert.Nil(t, err)
	assert.Equal(t, addr, sender.String())
}

func TestMultiSinger(t *testing.T) {
	InitEIP155Signer(big.NewInt(1))

	tx, err := GenerateDynamicFeeTxAndSinger()
	assert.Nil(t, err)
	err = tx.VerifySignature()
	assert.Nil(t, err)

	tx, err = GenerateAccessListTxAndSigner()
	assert.Nil(t, err)
	err = tx.VerifySignature()
	assert.Nil(t, err)
}

func TestLoadSignerWithPk(t *testing.T) {
	InitEIP155Signer(big.NewInt(1))
	s, err := LoadSignerWithPk("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	assert.Nil(t, err)
	assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", s.Addr.String())
}

func TestWrongSinger(t *testing.T) {
	InitEIP155Signer(big.NewInt(1))
	tx, s, err := GenerateWrongSignTransactionAndSigner(true)
	assert.Nil(t, err)

	addr := s.Addr
	err = tx.VerifySignature()
	assert.Contains(t, err.Error(), "verify signature failed")

	tx, s, err = GenerateWrongSignTransactionAndSigner(false)
	assert.Nil(t, err)
	sender, err := tx.sender()
	assert.Nil(t, err)
	assert.NotEqual(t, addr.String(), sender.String())
}

func TestEthTransaction_GetDynamicFeeSignHash(t *testing.T) {
	rawTx := "0x02f86d827a691e843b9aca08843b9aca088378d99c94450c8a57bae0aa50fa5122c84419d2b2924f205d0180c080a0b8fb5999b8ff73fd82b0b099dc3df62dbf3b9005115ee6122ff97acc01c9d39fa03e63ec1a91ae72246fd746a7e3a191d6c679f34c3358c6a01875a793f70d77bb"
	tx := &Transaction{}
	err := tx.Unmarshal(hexutil.Decode(rawTx))
	assert.Nil(t, err)

	addr := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	InitEIP155Signer(big.NewInt(31337))
	err = tx.VerifySignature()
	assert.Nil(t, err)

	sender, err := tx.sender()
	assert.Nil(t, err)
	assert.Equal(t, addr, sender.String())
}

func TestEthTransaction_Generate(t *testing.T) {
	InitEIP155Signer(big.NewInt(1))

	from, err := GenerateSigner()
	assert.Nil(t, err)
	tx, err := GenerateTransactionWithSigner(1, NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil, from)
	assert.Nil(t, err)
	decodeFrom, err := tx.sender()
	assert.Nil(t, err)
	assert.Equal(t, from.Addr.Bytes(), decodeFrom.Bytes())
	decodeFrom = tx.GetFrom()
	assert.Nil(t, err)
	assert.Equal(t, from.Addr.Bytes(), decodeFrom.Bytes())
}

// TestTransaction_PB is a test function that tests the serialization and deserialization
// of different types of transactions to and from Protocol Buffers.
func TestTransaction_PB(t *testing.T) {
	assert.True(t, common.IsHexAddress("0x264e23168e80f15e9311F2B88b4D7abeAba47E54"))
	addr := common.HexToAddress("0x264e23168e80f15e9311F2B88b4D7abeAba47E54")

	/** test case 1 LegacyTx */
	legacyTx := &Transaction{
		Inner: &LegacyTx{
			Nonce:    100,
			GasPrice: big.NewInt(5000000),
			Gas:      100000000,
			To:       &addr,
			Value:    big.NewInt(1),
			Data:     []byte{0, 1, 2, 3, 4, 5, 6, 7, 8},
			V:        big.NewInt(2),
			R:        big.NewInt(3),
			S:        big.NewInt(4),
		},
		Time: time.Now(),
		hash: atomic.Value{},
		size: atomic.Value{},
		from: atomic.Value{},
	}

	legacyTxPB := legacyTx.toPB()
	a, err := legacyTxPB.MarshalVT()
	assert.Nil(t, err)
	assert.NotNil(t, legacyTxPB)

	legacyTxPB2 := &pb.Transaction{}
	err = legacyTxPB2.UnmarshalVT(a)
	assert.Nil(t, err)

	legacyTx2 := &Transaction{}
	legacyTx2.fromPB(legacyTxPB2)
	// Ignore monotonic clock readings
	assert.Equal(t, legacyTx.Inner.GetGasPrice(), legacyTx2.Inner.GetGasPrice())
	if legacyTx2.Inner.GetTo() != nil {
		assert.Equal(t, legacyTx.Inner.GetTo().String(), legacyTx2.Inner.GetTo().String())
	}

	assert.True(t, legacyTx2.Time.Before(legacyTx.Time))
	/** test case 2 AccessListTx */
	accessListTx := &Transaction{
		Inner: &AccessListTx{
			Nonce:    100,
			GasPrice: big.NewInt(5000000),
			Gas:      100000000,
			To:       &addr,
			Value:    big.NewInt(1),
			Data:     []byte{0, 1, 2, 3, 4, 5, 6, 7, 8},
			V:        big.NewInt(2),
			R:        big.NewInt(3),
			S:        big.NewInt(4),
			AccessList: types.AccessList{
				{
					Address:     addr,
					StorageKeys: []common.Hash{common.BytesToHash([]byte{0}), common.BytesToHash([]byte{1})},
				},
			},
		},
		Time: time.Now(),
		hash: atomic.Value{},
		size: atomic.Value{},
	}

	accessListTxPB := accessListTx.toPB()
	assert.NotNil(t, accessListTxPB)

	accessListTx2 := &Transaction{}
	accessListTx2.fromPB(accessListTxPB)
	assert.Equal(t, accessListTx.Inner.GetGasPrice(), accessListTx2.Inner.GetGasPrice())
	assert.Equal(t, accessListTx.Inner.GetTo().String(), accessListTx2.Inner.GetTo().String())
	assert.Equal(t, accessListTx.Inner.GetAccessList()[0].Address.String(), accessListTx2.Inner.GetAccessList()[0].Address.String())
	assert.Equal(t, accessListTx.Inner.GetAccessList()[0].StorageKeys[0].String(), accessListTx2.Inner.GetAccessList()[0].StorageKeys[0].String())
	assert.True(t, accessListTx2.Time.Before(accessListTx.Time))

	/** test case 3 DynamicFeeTx */
	dynamicFeeTx := &Transaction{
		Inner: &DynamicFeeTx{
			Nonce:     100,
			GasTipCap: big.NewInt(5000000),
			GasFeeCap: big.NewInt(5000000),
			Gas:       100000000,
			To:        &addr,
			Value:     big.NewInt(1),
			Data:      []byte{0, 1, 2, 3, 4, 5, 6, 7, 8},
			V:         big.NewInt(2),
			R:         big.NewInt(3),
			S:         big.NewInt(4),
		},
		Time: time.Now(),
		hash: atomic.Value{},
		size: atomic.Value{},
	}

	dynamicFeeTxPB := dynamicFeeTx.toPB()
	assert.NotNil(t, dynamicFeeTxPB)

	dynamicFeeTx2 := &Transaction{}
	dynamicFeeTx2.fromPB(dynamicFeeTxPB)
	assert.Equal(t, dynamicFeeTx.Inner.GetGasTipCap(), dynamicFeeTx2.Inner.GetGasTipCap())
	assert.Equal(t, dynamicFeeTx.Inner.GetGasFeeCap(), dynamicFeeTx2.Inner.GetGasFeeCap())
	assert.Equal(t, dynamicFeeTx.Inner.GetTo().String(), dynamicFeeTx2.Inner.GetTo().String())
	assert.True(t, dynamicFeeTx2.Time.Before(dynamicFeeTx.Time))
}
