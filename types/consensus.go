package types

import (
	"math/big"
)

// TXConstraint is used to ensure that the pointer of T must be RbftTransaction
type TXConstraint[T any] interface {
	*T
	RbftTransaction
}

type RbftTransaction interface {
	RbftGetTxHash() string
	RbftGetFrom() string
	RbftGetTo() string
	RbftGetTimeStamp() int64
	RbftGetData() []byte
	RbftGetNonce() uint64
	RbftUnmarshal(raw []byte) error
	RbftMarshal() ([]byte, error)
	RbftIsConfigTx() bool
	RbftGetSize() int
	RbftGetGasPrice() *big.Int
	RbftGetGasLimit() uint64
	RbftGetGasFeeCap() *big.Int
	RbftGetValue() *big.Int
	RbftGetAccessList() AccessList
}
