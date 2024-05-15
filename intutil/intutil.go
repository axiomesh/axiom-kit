package intutil

import (
	"errors"
	"math/big"

	"github.com/holiman/uint256"
)

func BigIntToUint256(bigInt *big.Int) (*uint256.Int, error) {
	u256, overflow := uint256.FromBig(bigInt)
	if overflow {
		return nil, errors.New("overflow occurred while converting big.Int to uint256")
	}
	return u256, nil
}

func Uint256ToBigInt(u256 *uint256.Int) *big.Int {
	return new(big.Int).SetBytes(u256.Bytes())
}
