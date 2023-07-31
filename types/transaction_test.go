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
