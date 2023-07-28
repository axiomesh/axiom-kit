package types

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlock_Marshal(t *testing.T) {
	tests := []*Block{
		{},
		{
			BlockHeader:  &BlockHeader{},
			Transactions: []*Transaction{},
			BlockHash:    &Hash{},
			Signature:    []byte{},
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
