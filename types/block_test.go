package types

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBlock_Marshal(t *testing.T) {
	tests := []*Block{
		{},
		{
			BlockHeader:  &BlockHeader{},
			Transactions: []*Transaction{},
			BlockHash:    &Hash{},
			Extra:        []byte{},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			data, err := tt.Marshal()
			assert.Nil(t, err)

			e := &Block{}
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
	tests := []*Block{
		{},
		{
			BlockHeader:  &BlockHeader{},
			Transactions: []*Transaction{tx1, tx2},
			BlockHash:    &Hash{},
			Extra:        []byte{},
		},
		{
			BlockHeader: &BlockHeader{
				Number:          0,
				StateRoot:       NewHashByStr("1111111111111111111111111111111111111111111111111111111111111111"),
				TxRoot:          NewHashByStr("2222222222222222222222222222222222222222222222222222222222222222"),
				ReceiptRoot:     NewHashByStr("3333333333333333333333333333333333333333333333333333333333333333"),
				ParentHash:      NewHashByStr("4444444444444444444444444444444444444444444444444444444444444444"),
				Timestamp:       0,
				Epoch:           0,
				Bloom:           new(Bloom),
				GasPrice:        0,
				ProposerAccount: "111",
				GasUsed:         0,
				ProposerNodeID:  0,
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

			if tt.BlockHash != nil {
				assert.Equal(t, tt.BlockHash.String(), clonedBlock.BlockHash.String())
			}

			assert.Equal(t, tt.Extra, clonedBlock.Extra)

			if tt.BlockHeader != nil {
				assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader).Pointer())

				if tt.BlockHeader.TxRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader.TxRoot).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader.TxRoot).Pointer())
					assert.Equal(t, tt.BlockHeader.TxRoot.String(), clonedBlock.BlockHeader.TxRoot.String())
				}
				if tt.BlockHeader.StateRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader.StateRoot).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader.StateRoot).Pointer())
					assert.Equal(t, tt.BlockHeader.StateRoot.String(), clonedBlock.BlockHeader.StateRoot.String())
				}
				if tt.BlockHeader.ReceiptRoot != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader.ReceiptRoot).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader.ReceiptRoot).Pointer())
					assert.Equal(t, tt.BlockHeader.ReceiptRoot.String(), clonedBlock.BlockHeader.ReceiptRoot.String())
				}
				if tt.BlockHeader.Bloom != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader.Bloom).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader.Bloom).Pointer())
					assert.Equal(t, tt.BlockHeader.Bloom.Bytes(), clonedBlock.BlockHeader.Bloom.Bytes())
				}

				if tt.BlockHeader.ParentHash != nil {
					assert.NotEqual(t, reflect.ValueOf(tt.BlockHeader.ParentHash).Pointer(), reflect.ValueOf(clonedBlock.BlockHeader.ParentHash).Pointer())
					assert.Equal(t, tt.BlockHeader.ParentHash.String(), clonedBlock.BlockHeader.ParentHash.String())
				}

				assert.Equal(t, tt.BlockHeader.Number, clonedBlock.BlockHeader.Number)
				assert.Equal(t, tt.BlockHeader.Epoch, clonedBlock.BlockHeader.Epoch)
				assert.Equal(t, tt.BlockHeader.GasPrice, clonedBlock.BlockHeader.GasPrice)
				assert.Equal(t, tt.BlockHeader.GasUsed, clonedBlock.BlockHeader.GasUsed)
				assert.Equal(t, tt.BlockHeader.Timestamp, clonedBlock.BlockHeader.Timestamp)
				assert.Equal(t, tt.BlockHeader.ProposerAccount, clonedBlock.BlockHeader.ProposerAccount)
				assert.Equal(t, tt.BlockHeader.ProposerNodeID, clonedBlock.BlockHeader.ProposerNodeID)
			}
		})
	}
}
