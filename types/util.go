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
	factorPart := big.NewInt(1)
	var uInfo *unitInfo
	for _, info := range unitMap {
		if strings.HasSuffix(valueStr, info.Str) {
			numericPart = strings.TrimSpace(strings.TrimSuffix(valueStr, info.Str))
			factorPart = info.Num
			uInfo = info
			break
		}
	}

	numericValue, ok := new(big.Float).SetString(numericPart)
	if !ok {
		return nil, fmt.Errorf("invalid coin number format: %v, only support mol/axc/gmol uint, e.g: 100mol, 1.1axc, 1000.1gmol", valueStr)
	}

	decimalPart := strings.Split(numericPart, ".")
	if len(decimalPart) == 2 {
		if uInfo == nil {
			if len(decimalPart[1]) > 0 {
				return nil, fmt.Errorf("invalid coin number format: %v, must be an integer wihout unit", valueStr)
			}
		} else {
			if len(decimalPart[1]) > uInfo.Decimals {
				return nil, fmt.Errorf("invalid coin number format: %v, decimal part length must be less or equal than %v", valueStr, uInfo.Decimals)
			}
		}
	}

	numericValue = new(big.Float).Mul(numericValue, new(big.Float).SetInt(factorPart))
	res := new(big.Int)
	res, _ = numericValue.Int(res)
	return (*CoinNumber)(res), nil
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
	return (*big.Int)(c)
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
