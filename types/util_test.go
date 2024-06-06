package types

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCoinNumber(t *testing.T) {
	tests := []struct {
		valueStr string
		want     func() *CoinNumber
		wantErr  assert.ErrorAssertionFunc
	}{
		// no error
		{valueStr: "1", want: func() *CoinNumber {
			return CoinNumberByMol(1)
		}, wantErr: assert.NoError},
		{valueStr: "1mol", want: func() *CoinNumber {
			return CoinNumberByMol(1)
		}, wantErr: assert.NoError},
		{valueStr: "1 mol", want: func() *CoinNumber {
			return CoinNumberByMol(1)
		}, wantErr: assert.NoError},
		{valueStr: "1gmol", want: func() *CoinNumber {
			return CoinNumberByGmol(1)
		}, wantErr: assert.NoError},
		{valueStr: "1 gmol", want: func() *CoinNumber {
			return CoinNumberByGmol(1)
		}, wantErr: assert.NoError},
		{valueStr: "1.111111111gmol", want: func() *CoinNumber {
			return CoinNumberByMol(1_100_000_000)
		}, wantErr: assert.NoError},
		{valueStr: "1.111111111 gmol", want: func() *CoinNumber {
			return CoinNumberByMol(1_100_000_000)
		}, wantErr: assert.NoError},
		{valueStr: "1axc", want: func() *CoinNumber {
			return CoinNumberByAxc(1)
		}, wantErr: assert.NoError},
		{valueStr: "1 axc", want: func() *CoinNumber {
			return CoinNumberByAxc(1)
		}, wantErr: assert.NoError},
		{valueStr: "1.111111111axc", want: func() *CoinNumber {
			return CoinNumberByGmol(1_100_000_000)
		}, wantErr: assert.NoError},
		{valueStr: "1.111111111 axc", want: func() *CoinNumber {
			return CoinNumberByGmol(1_100_000_000)
		}, wantErr: assert.NoError},
		{valueStr: "100001576744488228294", want: func() *CoinNumber {
			c, _ := new(big.Int).SetString("100001576744488228294", 10)
			return CoinNumberByBigInt(c)
		}, wantErr: assert.NoError},

		// error
		{valueStr: "1.1", want: nil, wantErr: assert.Error},
		{valueStr: "1.1 axc2", want: nil, wantErr: assert.Error},
		{valueStr: "1.1mol", want: nil, wantErr: assert.Error},
		// accuracy overflow
		{valueStr: "1.1111111111gmol", want: nil, wantErr: assert.Error},
		{valueStr: "1.1111111111111111111axc", want: nil, wantErr: assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.valueStr, func(t *testing.T) {
			got, err := ParseCoinNumber(tt.valueStr)
			if tt.wantErr(t, err, fmt.Sprintf("ParseCoinNumber(%v)", tt.valueStr)) {
				return
			}
			assert.Truef(t, tt.want().ToBigInt().Cmp(got.ToBigInt()) == 0, "ParseCoinNumber(%v)", tt.valueStr)
		})
	}
}

func TestCoinNumber_String(t *testing.T) {
	tests := []struct {
		coinNumber *CoinNumber
		want       string
	}{
		{coinNumber: CoinNumberByMol(1), want: "1mol"},
		{coinNumber: CoinNumberByGmol(1), want: "1gmol"},
		{coinNumber: CoinNumberByAxc(1), want: "1axc"},
		{coinNumber: CoinNumberByMol(999999999), want: "999999999mol"},
		{coinNumber: CoinNumberByMol(1_999999999), want: "1.999999999gmol"},
		{coinNumber: CoinNumberByGmol(999999999), want: "999999999gmol"},
		{coinNumber: CoinNumberByGmol(1_999999999), want: "1.999999999axc"},
		{coinNumber: CoinNumberByMol(1_999999999_999999999), want: "1.999999999999999999axc"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.coinNumber.String(), "coinNumber(%v)", tt.want)
		})
	}
}
