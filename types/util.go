package types

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

var MolUnitStr = "mol"
var GmolUnitStr = "gmol"
var AxcUnitStr = "axc"

var MolUnit = big.NewInt(1)
var GmolUnit = big.NewInt(1_000000000)
var AxcUnit = big.NewInt(1_000000000_000000000)

type unitInfo struct {
	Str      string
	Num      *big.Int
	Decimals int
}

var unitMap = []*unitInfo{
	{
		MolUnitStr,
		MolUnit,
		0,
	},
	{
		GmolUnitStr,
		GmolUnit,
		9,
	},
	{
		AxcUnitStr,
		AxcUnit,
		18,
	},
}

func init() {
	sort.Slice(unitMap, func(i, j int) bool {
		return len(unitMap[i].Str) > len(unitMap[j].Str)
	})
}

type CoinNumber big.Int

func CoinNumberByBigInt(x *big.Int) *CoinNumber {
	return (*CoinNumber)(new(big.Int).Set(x))
}

func CoinNumberByMol(mol uint64) *CoinNumber {
	return (*CoinNumber)(big.NewInt(0).SetUint64(mol))
}

func CoinNumberByGmol(gmol uint64) *CoinNumber {
	return (*CoinNumber)(new(big.Int).Mul(new(big.Int).SetUint64(gmol), GmolUnit))
}

func CoinNumberByAxc(axc uint64) *CoinNumber {
	return (*CoinNumber)(new(big.Int).Mul(new(big.Int).SetUint64(axc), AxcUnit))
}

func ParseCoinNumber(valueStr string) (*CoinNumber, error) {
	valueStr = strings.TrimSpace(valueStr)
	if valueStr == "" {
		return new(CoinNumber), nil
	}

	numericPart := valueStr
	decimals := 0
	for _, info := range unitMap {
		if strings.HasSuffix(valueStr, info.Str) {
			numericPart = strings.TrimSpace(strings.TrimSuffix(valueStr, info.Str))
			decimals = info.Decimals
			break
		}
	}

	decimalPart := strings.Split(numericPart, ".")
	intPart := decimalPart[0]
	decPart := ""
	if len(decimalPart) == 2 {
		decPart = decimalPart[1]
		if len(decPart) > decimals {
			return nil, fmt.Errorf("invalid coin number format: %v, decimal part length must be less or equal than %v", valueStr, decimals)
		}
	}

	fullNumber := intPart + decPart + strings.Repeat("0", decimals-len(decPart))
	if fullNumber == "" {
		fullNumber = "0"
	}

	numericValue := new(big.Int)
	numericValue, ok := numericValue.SetString(fullNumber, 10)
	if !ok {
		return nil, fmt.Errorf("invalid numeric value: %v", fullNumber)
	}

	return (*CoinNumber)(numericValue), nil
}

func (c *CoinNumber) MarshalText() (text []byte, err error) {
	return []byte(c.String()), nil
}

func (c *CoinNumber) UnmarshalText(b []byte) error {
	x, err := ParseCoinNumber(string(b))
	if err != nil {
		return err
	}
	*c = *x
	return nil
}

func (c *CoinNumber) ToBigInt() *big.Int {
	return new(big.Int).Set((*big.Int)(c))
}

func (c *CoinNumber) String() string {
	bigInt := c.ToBigInt()
	bigIntStr := bigInt.String()
	if bigInt.Cmp(GmolUnit) < 0 {
		return fmt.Sprintf("%s%s", bigIntStr, MolUnitStr)
	}
	if bigInt.Cmp(AxcUnit) < 0 {
		return fmt.Sprintf("%s%s", insertDot(bigIntStr, 9), GmolUnitStr)
	}
	return fmt.Sprintf("%s%s", insertDot(bigIntStr, 18), AxcUnitStr)
}

func (c *CoinNumber) InitNoZero() {
	one := CoinNumberByMol(1)
	*c = *one
}

func (c *CoinNumber) Clone() *CoinNumber {
	if c == nil {
		return nil
	}
	return (*CoinNumber)(new(big.Int).Set(c.ToBigInt()))
}

func insertDot(str string, count int) string {
	insertIndex := len(str) - count
	res := str[:insertIndex] + "." + str[insertIndex:]
	res = strings.TrimRight(res, "0")
	res = strings.TrimRight(res, ".")
	return res
}

func IsZeroBytes(bytes []byte) bool {
	b := byte(0)
	for _, s := range bytes {
		b |= s
	}
	return b == 0
}
