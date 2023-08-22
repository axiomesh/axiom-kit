package types

import (
	"math/big"
	"testing"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestEthTransaction_GetSignHash(t *testing.T) {
	rawTx := "0xf86c8085147d35700082520894f927bb571eaab8c9a361ab405c9e4891c5024380880de0b6b3a76400008025a00b8e3b66c1e7ae870802e3ef75f1ec741f19501774bd5083920ce181c2140b99a0040c122b7ebfb3d33813927246cbbad1c6bf210474f5d28053990abff0fd4f53"
	tx := &Transaction{}
	tx.Unmarshal(hexutil.Decode(rawTx))

	addr := "0xC63573cB77ec56e0A1cb40199bb85838D71e4dce"

	InitEIP155Signer(big.NewInt(1))

	err := tx.VerifySignature()
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
	tx.Unmarshal(hexutil.Decode(rawTx))

	addr := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	InitEIP155Signer(big.NewInt(31337))
	err := tx.VerifySignature()
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
