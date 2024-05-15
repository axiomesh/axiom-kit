package types

import (
	"math/big"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBlock_Marshal(t *testing.T) {
	from, err := GenerateSigner()
	assert.Nil(t, err)
	tx, err := GenerateTransactionWithSigner(1, NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil, from)
	assert.Nil(t, err)

	tests := []*Block{
		{},
		{
			Header:       &BlockHeader{},
			Transactions: []*Transaction{},
			Extra: &BlockExtra{
				Size: 1,
			},
		},
		{
			Header: &BlockHeader{
				Number:         0,
				StateRoot:      NewHashByStr("0x416cff32e16ee5b024d89a5ec055a706d8f9d289859f7e1f9622610b13f8067c"),
				TxRoot:         NewHashByStr("0x416cff32e16ee5b024d89a5ec055a706d8f9d289859f7e1f9622610b13f8067c"),
				ReceiptRoot:    NewHashByStr("0x416cff32e16ee5b024d89a5ec055a706d8f9d289859f7e1f9622610b13f8067c"),
				ParentHash:     NewHashByStr("0x416cff32e16ee5b024d89a5ec055a706d8f9d289859f7e1f9622610b13f8067c"),
				Timestamp:      0,
				Epoch:          0,
				Bloom:          new(Bloom),
				GasPrice:       0,
				GasUsed:        0,
				ProposerNodeID: 0,
			},
			Transactions: []*Transaction{
				tx,
			}},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			data, err := tt.Marshal()
			assert.Nil(t, err)

			e := &Block{}
			err = e.Unmarshal(data)
			assert.Nil(t, err)
			assert.Equal(t, tt.Hash(), e.Hash())
		})
	}
}

func TestBlockBody_Marshal(t *testing.T) {
	from, err := GenerateSigner()
	assert.Nil(t, err)
	tx, err := GenerateTransactionWithSigner(1, NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(1), nil, from)
	assert.Nil(t, err)

	tests := []*BlockBody{
		{},
		{
			Transactions: []*Transaction{},
			Extra: &BlockExtra{
				Size: 1,
			},
		},
		{
			Transactions: []*Transaction{
				tx,
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			data, err := tt.Marshal()
			assert.Nil(t, err)

			e := &BlockBody{}
			err = e.Unmarshal(data)
			assert.Nil(t, err)
		})
	}
}

func TestClone(t *testing.T) {
	tx1, err := GenerateEmptyTransactionAndSigner()
	assert.Nil(t, err)
	tx2, err := GenerateEmptyTransactionAndSigner()
	assert.Nil(t, err)
	bloom := new(Bloom)
	bloom[0] = 1
	tests := []*Block{
		{
			Header:       &BlockHeader{},
			Transactions: []*Transaction{tx1, tx2},
			Extra:        &BlockExtra{},
		},
		{},
		{
			Header: &BlockHeader{
				Number:         1,
				StateRoot:      NewHashByStr("1111111111111111111111111111111111111111111111111111111111111111"),
				TxRoot:         NewHashByStr("2222222222222222222222222222222222222222222222222222222222222222"),
				ReceiptRoot:    NewHashByStr("3333333333333333333333333333333333333333333333333333333333333333"),
				ParentHash:     NewHashByStr("4444444444444444444444444444444444444444444444444444444444444444"),
				Timestamp:      1,
				Epoch:          1,
				Bloom:          bloom,
				ProposerNodeID: 1,
				GasPrice:       1,
				GasUsed:        1,
				TotalGasFee:    big.NewInt(1),
				GasFeeReward:   big.NewInt(1),
				hashCache:      atomic.Value{},
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			clonedBlock := tt.Clone()
			assert.NotEqual(t, reflect.ValueOf(tt).Pointer(), reflect.ValueOf(clonedBlock).Pointer())

			if len(tt.Transactions) != 0 {
				assert.Equal(t, len(tt.Transactions), len(clonedBlock.Transactions))
				lo.ForEach(tt.Transactions, func(expectTx *Transaction, i int) {
					assert.NotEqual(t, reflect.ValueOf(expectTx).Pointer(), reflect.ValueOf(clonedBlock.Transactions[i]).Pointer())
					expectData, err := expectTx.MarshalBinary()
					assert.Nil(t, err)
					actualData, err := clonedBlock.Transactions[i].MarshalBinary()
					assert.Nil(t, err)
					assert.Equal(t, expectData, actualData)
				})
			}

			if tt.Header != nil {
				assert.NotEmpty(t, tt.Hash().String())
				assert.Equal(t, tt.Hash().String(), clonedBlock.Hash().String())
			}

			assert.True(t, reflect.DeepEqual(tt.Extra, clonedBlock.Extra))

			if tt.Header != nil {
				assert.NotEqual(t, reflect.ValueOf(tt.Header).Pointer(), reflect.ValueOf(clonedBlock.Header).Pointer())

				if tt.Header.TxRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.Header.TxRoot).Pointer(), reflect.ValueOf(clonedBlock.Header.TxRoot).Pointer())
					assert.Equal(t, tt.Header.TxRoot.String(), clonedBlock.Header.TxRoot.String())
				}
				if tt.Header.StateRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.Header.StateRoot).Pointer(), reflect.ValueOf(clonedBlock.Header.StateRoot).Pointer())
					assert.Equal(t, tt.Header.StateRoot.String(), clonedBlock.Header.StateRoot.String())
				}
				if tt.Header.ReceiptRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.Header.ReceiptRoot).Pointer(), reflect.ValueOf(clonedBlock.Header.ReceiptRoot).Pointer())
					assert.Equal(t, tt.Header.ReceiptRoot.String(), clonedBlock.Header.ReceiptRoot.String())
				}
				if tt.Header.Bloom != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.Header.Bloom).Pointer(), reflect.ValueOf(clonedBlock.Header.Bloom).Pointer())
					assert.Equal(t, tt.Header.Bloom.Bytes(), clonedBlock.Header.Bloom.Bytes())
				}

				if tt.Header.ParentHash != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.Header.ParentHash).Pointer(), reflect.ValueOf(clonedBlock.Header.ParentHash).Pointer())
					assert.Equal(t, tt.Header.ParentHash.String(), clonedBlock.Header.ParentHash.String())
				}

				assert.Equal(t, tt.Header.Number, clonedBlock.Header.Number)
				assert.Equal(t, tt.Header.Epoch, clonedBlock.Header.Epoch)
				assert.Equal(t, tt.Header.GasPrice, clonedBlock.Header.GasPrice)
				assert.Equal(t, tt.Header.GasUsed, clonedBlock.Header.GasUsed)
				assert.Equal(t, tt.Header.Timestamp, clonedBlock.Header.Timestamp)
				assert.Equal(t, tt.Header.ProposerNodeID, clonedBlock.Header.ProposerNodeID)
				assert.Equal(t, tt.Header.TotalGasFee, clonedBlock.Header.TotalGasFee)
				assert.Equal(t, tt.Header.GasFeeReward, clonedBlock.Header.GasFeeReward)
			}
		})
	}
}
