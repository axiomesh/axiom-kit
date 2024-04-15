package intutil

import (
	"fmt"
	"github.com/holiman/uint256"
	"math/big"
)

func BigIntToUint256(bigInt *big.Int) (*uint256.Int, error) {
	u256, overflow := uint256.FromBig(bigInt)
	if overflow {
		return nil, fmt.Errorf("overflow occurred while converting big.Int to uint256")
	}
	return u256, nil
}

func Uint256ToBigInt(u256 *uint256.Int) *big.Int {
	return new(big.Int).SetBytes(u256.Bytes())
}
