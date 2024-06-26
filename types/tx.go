package types

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/axiomesh/axiom-kit/types/pb"
)

// Transaction types.
const (
	LegacyTxType = iota
	AccessListTxType
	DynamicFeeTxType
	IncentiveTxType
)

// TxData is the underlying data of a transaction.
//
// This is implemented by DynamicFeeTx, LegacyTx and AccessListTx.
type TxData interface {
	TxType() byte // returns the type ID
	copy() TxData // creates a deep copy and initializes all fields

	GetChainID() *big.Int
	GetAccessList() types.AccessList
	GetData() []byte
	GetGas() uint64
	GetGasPrice() *big.Int
	GetGasTipCap() *big.Int
	GetGasFeeCap() *big.Int
	GetValue() *big.Int
	GetNonce() uint64
	GetTo() *common.Address
	GetIncentiveAddress() *common.Address

	RawSignatureValues() (v, r, s *big.Int)
	setSignatureValues(chainID, v, r, s *big.Int)
	EffectiveGasPrice(baseFee *big.Int) *big.Int
}

// AccessListTx is the data of EIP-2930 access list transactions.
type AccessListTx struct {
	ChainID    *big.Int         // destination chain ID
	Nonce      uint64           // nonce of sender account
	GasPrice   *big.Int         // wei per gas
	Gas        uint64           // gas limit
	To         *common.Address  `rlp:"nil"` // nil means contract creation
	Value      *big.Int         // wei amount
	Data       []byte           // contract invocation input data
	AccessList types.AccessList // EIP-2930 access list
	V, R, S    *big.Int         // signature values
}

// LegacyTx is the transaction data of regular Ethereum transactions.
type LegacyTx struct {
	Nonce    uint64          // nonce of sender account
	GasPrice *big.Int        // wei per gas
	Gas      uint64          // gas limit
	To       *common.Address `rlp:"nil"` // nil means contract creation
	Value    *big.Int        // wei amount
	Data     []byte          // contract invocation input data
	V, R, S  *big.Int        // signature values
}

type DynamicFeeTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	Gas        uint64
	To         *common.Address `rlp:"nil"` // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList types.AccessList

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`

	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

type IncentiveTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	Gas        uint64
	To         *common.Address `rlp:"nil"` // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList types.AccessList

	IncentiveAddress *common.Address

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`

	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *AccessListTx) copy() TxData {
	cpy := &AccessListTx{
		Nonce: tx.Nonce,
		To:    CopyAddress(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(types.AccessList, len(tx.AccessList)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasPrice:   new(big.Int),
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasPrice != nil {
		cpy.GasPrice.Set(tx.GasPrice)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.

func (tx *AccessListTx) TxType() byte { return AccessListTxType }

func (tx *AccessListTx) GetChainID() *big.Int { return tx.ChainID }

func (tx *AccessListTx) protected() bool { return true }

func (tx *AccessListTx) GetAccessList() types.AccessList { return tx.AccessList }

func (tx *AccessListTx) GetData() []byte { return tx.Data }

func (tx *AccessListTx) GetGas() uint64 { return tx.Gas }

func (tx *AccessListTx) GetGasPrice() *big.Int { return tx.GasPrice }

func (tx *AccessListTx) GetGasTipCap() *big.Int { return tx.GasPrice }

func (tx *AccessListTx) GetGasFeeCap() *big.Int { return tx.GasPrice }

func (tx *AccessListTx) GetValue() *big.Int { return tx.Value }

func (tx *AccessListTx) GetNonce() uint64 { return tx.Nonce }

func (tx *AccessListTx) GetTo() *common.Address { return tx.To }

func (tx *AccessListTx) GetIncentiveAddress() *common.Address {
	return nil
}

func (tx *AccessListTx) RawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *AccessListTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *AccessListTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return new(big.Int).Set(tx.GasPrice)
}

func (tx *AccessListTx) toPB() *pb.AccessListTx {
	if tx == nil {
		return &pb.AccessListTx{}
	}

	return &pb.AccessListTx{
		ChainId:    toPBBigInt(tx.ChainID),
		Nonce:      tx.Nonce,
		GasPrice:   toPBBigInt(tx.GasPrice),
		Gas:        tx.Gas,
		To:         toPBAddress(tx.To),
		Value:      toPBBigInt(tx.Value),
		Data:       tx.Data,
		AccessList: toPBAccessList(tx.AccessList),
		V:          toPBBigInt(tx.V),
		R:          toPBBigInt(tx.R),
		S:          toPBBigInt(tx.S),
	}
}

func (tx *AccessListTx) fromPB(pb *pb.AccessListTx) {
	tx.ChainID = fromPBBigInt(pb.ChainId)
	tx.Nonce = pb.Nonce
	tx.GasPrice = fromPBBigInt(pb.GasPrice)
	tx.Gas = pb.Gas
	tx.To = fromPBAddress(pb.To)
	tx.Value = fromPBBigInt(pb.Value)
	tx.Data = pb.Data
	tx.AccessList = fromPBAccessList(pb.AccessList)
	tx.V = fromPBBigInt(pb.V)
	tx.R = fromPBBigInt(pb.R)
	tx.S = fromPBBigInt(pb.S)
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *LegacyTx) copy() TxData {
	cpy := &LegacyTx{
		Nonce: tx.Nonce,
		To:    CopyAddress(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are initialized below.
		Value:    new(big.Int),
		GasPrice: new(big.Int),
		V:        new(big.Int),
		R:        new(big.Int),
		S:        new(big.Int),
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.GasPrice != nil {
		cpy.GasPrice.Set(tx.GasPrice)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.

func (tx *LegacyTx) TxType() byte { return LegacyTxType }

func (tx *LegacyTx) GetChainID() *big.Int { return deriveChainId(tx.V) }

func (tx *LegacyTx) GetAccessList() types.AccessList { return nil }

func (tx *LegacyTx) GetData() []byte { return tx.Data }

func (tx *LegacyTx) GetGas() uint64 { return tx.Gas }

func (tx *LegacyTx) GetGasPrice() *big.Int { return tx.GasPrice }

func (tx *LegacyTx) GetGasTipCap() *big.Int { return tx.GasPrice }

func (tx *LegacyTx) GetGasFeeCap() *big.Int { return tx.GasPrice }

func (tx *LegacyTx) GetValue() *big.Int { return tx.Value }

func (tx *LegacyTx) GetNonce() uint64 { return tx.Nonce }

func (tx *LegacyTx) GetTo() *common.Address { return tx.To }

func (tx *LegacyTx) GetIncentiveAddress() *common.Address {
	return nil
}

func (tx *LegacyTx) RawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *LegacyTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.V, tx.R, tx.S = v, r, s
}

func (tx *LegacyTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return new(big.Int).Set(tx.GasPrice)
}

func (tx *LegacyTx) toPB() *pb.LegacyTx {
	if tx == nil {
		return &pb.LegacyTx{}
	}

	return &pb.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: toPBBigInt(tx.GasPrice),
		Gas:      tx.Gas,
		To:       toPBAddress(tx.To),
		Value:    toPBBigInt(tx.Value),
		Data:     tx.Data,
		V:        toPBBigInt(tx.V),
		R:        toPBBigInt(tx.R),
		S:        toPBBigInt(tx.S),
	}
}

func (tx *LegacyTx) fromPB(pb *pb.LegacyTx) {
	tx.Nonce = pb.Nonce
	tx.GasPrice = fromPBBigInt(pb.GasPrice)
	tx.Gas = pb.Gas
	tx.To = fromPBAddress(pb.To)
	tx.Value = fromPBBigInt(pb.Value)
	tx.Data = pb.Data
	tx.V = fromPBBigInt(pb.V)
	tx.R = fromPBBigInt(pb.R)
	tx.S = fromPBBigInt(pb.S)
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *DynamicFeeTx) copy() TxData {
	cpy := &DynamicFeeTx{
		Nonce: tx.Nonce,
		To:    CopyAddress(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(types.AccessList, len(tx.AccessList)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *DynamicFeeTx) TxType() byte { return DynamicFeeTxType }

func (tx *DynamicFeeTx) GetChainID() *big.Int { return tx.ChainID }

func (tx *DynamicFeeTx) protected() bool { return true }

func (tx *DynamicFeeTx) GetAccessList() types.AccessList { return tx.AccessList }

func (tx *DynamicFeeTx) GetData() []byte { return tx.Data }

func (tx *DynamicFeeTx) GetGas() uint64 { return tx.Gas }

func (tx *DynamicFeeTx) GetGasFeeCap() *big.Int { return tx.GasFeeCap }

func (tx *DynamicFeeTx) GetGasTipCap() *big.Int { return tx.GasTipCap }

func (tx *DynamicFeeTx) GetGasPrice() *big.Int { return tx.GasFeeCap }

func (tx *DynamicFeeTx) GetValue() *big.Int { return tx.Value }

func (tx *DynamicFeeTx) GetNonce() uint64 { return tx.Nonce }

func (tx *DynamicFeeTx) GetTo() *common.Address { return tx.To }

func (tx *DynamicFeeTx) GetIncentiveAddress() *common.Address {
	return nil
}

func (tx *DynamicFeeTx) RawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *DynamicFeeTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *DynamicFeeTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	dst := new(big.Int)
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *DynamicFeeTx) toPB() *pb.DynamicFeeTx {
	if tx == nil {
		return &pb.DynamicFeeTx{}
	}

	return &pb.DynamicFeeTx{
		ChainId:    toPBBigInt(tx.ChainID),
		Nonce:      tx.Nonce,
		GasTipCap:  toPBBigInt(tx.GasTipCap),
		GasFeeCap:  toPBBigInt(tx.GasFeeCap),
		Gas:        tx.Gas,
		To:         toPBAddress(tx.To),
		Value:      toPBBigInt(tx.Value),
		Data:       tx.Data,
		AccessList: toPBAccessList(tx.AccessList),
		V:          toPBBigInt(tx.V),
		R:          toPBBigInt(tx.R),
		S:          toPBBigInt(tx.S),
	}
}

func (tx *DynamicFeeTx) fromPB(pb *pb.DynamicFeeTx) {
	tx.ChainID = fromPBBigInt(pb.ChainId)
	tx.Nonce = pb.Nonce
	tx.GasTipCap = fromPBBigInt(pb.GasTipCap)
	tx.GasFeeCap = fromPBBigInt(pb.GasFeeCap)
	tx.Gas = pb.Gas
	tx.To = fromPBAddress(pb.To)
	tx.Value = fromPBBigInt(pb.Value)
	tx.Data = pb.Data
	tx.AccessList = fromPBAccessList(pb.AccessList)
	tx.V = fromPBBigInt(pb.V)
	tx.R = fromPBBigInt(pb.R)
	tx.S = fromPBBigInt(pb.S)
}

func (tx *IncentiveTx) copy() TxData {
	cpy := &IncentiveTx{
		Nonce: tx.Nonce,
		To:    CopyAddress(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList:       make(types.AccessList, len(tx.AccessList)),
		Value:            new(big.Int),
		ChainID:          new(big.Int),
		GasTipCap:        new(big.Int),
		GasFeeCap:        new(big.Int),
		IncentiveAddress: CopyAddress(tx.IncentiveAddress),
		V:                new(big.Int),
		R:                new(big.Int),
		S:                new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

func (tx *IncentiveTx) TxType() byte { return IncentiveTxType }

func (tx *IncentiveTx) GetChainID() *big.Int { return tx.ChainID }

func (tx *IncentiveTx) protected() bool { return true }

func (tx *IncentiveTx) GetAccessList() types.AccessList { return tx.AccessList }

func (tx *IncentiveTx) GetData() []byte { return tx.Data }

func (tx *IncentiveTx) GetGas() uint64 { return tx.Gas }

func (tx *IncentiveTx) GetGasFeeCap() *big.Int { return tx.GasFeeCap }

func (tx *IncentiveTx) GetGasTipCap() *big.Int { return tx.GasTipCap }

func (tx *IncentiveTx) GetGasPrice() *big.Int { return tx.GasFeeCap }

func (tx *IncentiveTx) GetValue() *big.Int { return tx.Value }

func (tx *IncentiveTx) GetNonce() uint64 { return tx.Nonce }

func (tx *IncentiveTx) GetTo() *common.Address { return tx.To }

func (tx *IncentiveTx) GetIncentiveAddress() *common.Address {
	return tx.IncentiveAddress
}

func (tx *IncentiveTx) RawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *IncentiveTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *IncentiveTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	dst := new(big.Int)
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *IncentiveTx) toPB() *pb.IncentiveTx {
	if tx == nil {
		return &pb.IncentiveTx{}
	}

	return &pb.IncentiveTx{
		ChainId:          toPBBigInt(tx.ChainID),
		Nonce:            tx.Nonce,
		GasTipCap:        toPBBigInt(tx.GasTipCap),
		GasFeeCap:        toPBBigInt(tx.GasFeeCap),
		Gas:              tx.Gas,
		To:               toPBAddress(tx.To),
		Value:            toPBBigInt(tx.Value),
		Data:             tx.Data,
		AccessList:       toPBAccessList(tx.AccessList),
		IncentiveAddress: toPBAddress(tx.IncentiveAddress),
		V:                toPBBigInt(tx.V),
		R:                toPBBigInt(tx.R),
		S:                toPBBigInt(tx.S),
	}
}

func (tx *IncentiveTx) fromPB(pb *pb.IncentiveTx) {
	tx.ChainID = fromPBBigInt(pb.ChainId)
	tx.Nonce = pb.Nonce
	tx.GasTipCap = fromPBBigInt(pb.GasTipCap)
	tx.GasFeeCap = fromPBBigInt(pb.GasFeeCap)
	tx.Gas = pb.Gas
	tx.To = fromPBAddress(pb.To)
	tx.Value = fromPBBigInt(pb.Value)
	tx.Data = pb.Data
	tx.AccessList = fromPBAccessList(pb.AccessList)
	tx.IncentiveAddress = fromPBAddress(pb.IncentiveAddress)
	tx.V = fromPBBigInt(pb.V)
	tx.R = fromPBBigInt(pb.R)
	tx.S = fromPBBigInt(pb.S)
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

func RecoverPlain(hash []byte, R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return nil, errors.New("invalid signature")
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return nil, errors.New("invalid signature")
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}

	return crypto.Keccak256(pub[1:])[12:], nil
}

// RlpHash encodes x and hashes the encoded bytes.
func RlpHash(x any) *common.Hash {
	var h common.Hash
	sha := hasherPool.Get().(crypto.KeccakState)
	defer hasherPool.Put(sha)
	sha.Reset()
	rlp.Encode(sha, x)
	sha.Read(h[:])
	return &h
}

// PrefixedRlpHash writes the prefix into the hasher before rlp-encoding x.
// It's used for typed transactions.
func PrefixedRlpHash(prefix byte, x any) *common.Hash {
	var h common.Hash
	sha := hasherPool.Get().(crypto.KeccakState)
	defer hasherPool.Put(sha)
	sha.Reset()
	sha.Write([]byte{prefix})
	rlp.Encode(sha, x)
	sha.Read(h[:])
	return &h
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From                 *common.Address   `json:"from"`
	To                   *common.Address   `json:"to"`
	Gas                  *hexutil.Uint64   `json:"gas"`
	GasPrice             *hexutil.Big      `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big      `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big      `json:"maxPriorityFeePerGas"`
	Value                *hexutil.Big      `json:"value"`
	Nonce                *hexutil.Uint64   `json:"nonce"`
	Data                 *hexutil.Bytes    `json:"data"`
	Input                *hexutil.Bytes    `json:"input"`
	AccessList           *types.AccessList `json:"accessList"`
	ChainID              *hexutil.Big      `json:"chainId,omitempty"`
}

// GetFrom retrieves the transaction sender address.
func (args *CallArgs) GetFrom() common.Address {
	if args.From == nil {
		return common.Address{}
	}
	return *args.From
}

// GetData retrieves the transaction calldata. Input field is preferred.
func (args *CallArgs) GetData() []byte {
	if args.Input != nil {
		return *args.Input
	}
	if args.Data != nil {
		return *args.Data
	}
	return nil
}

// toTransaction converts the arguments to a transaction.
// This assumes that setDefaults has been called.
func (args *CallArgs) toTransaction() *types.Transaction {
	var data types.TxData
	switch {
	case args.MaxFeePerGas != nil:
		al := types.AccessList{}
		if args.AccessList != nil {
			al = *args.AccessList
		}
		data = &types.DynamicFeeTx{
			To:         args.To,
			ChainID:    (*big.Int)(args.ChainID),
			Nonce:      uint64(*args.Nonce),
			Gas:        uint64(*args.Gas),
			GasFeeCap:  (*big.Int)(args.MaxFeePerGas),
			GasTipCap:  (*big.Int)(args.MaxPriorityFeePerGas),
			Value:      (*big.Int)(args.Value),
			Data:       args.GetData(),
			AccessList: al,
		}
	case args.AccessList != nil:
		data = &types.AccessListTx{
			To:         args.To,
			ChainID:    (*big.Int)(args.ChainID),
			Nonce:      uint64(*args.Nonce),
			Gas:        uint64(*args.Gas),
			GasPrice:   (*big.Int)(args.GasPrice),
			Value:      (*big.Int)(args.Value),
			Data:       args.GetData(),
			AccessList: *args.AccessList,
		}
	default:
		data = &types.LegacyTx{
			To:       args.To,
			Nonce:    uint64(*args.Nonce),
			Gas:      uint64(*args.Gas),
			GasPrice: (*big.Int)(args.GasPrice),
			Value:    (*big.Int)(args.Value),
			Data:     args.GetData(),
		}
	}
	return types.NewTx(data)
}

// ToTransaction converts the arguments to a transaction.
// This assumes that setDefaults has been called.
func (args *CallArgs) ToTransaction() *types.Transaction {
	return args.toTransaction()
}

func CopyAddress(src *common.Address) *common.Address {
	if src == nil {
		return nil
	}

	copiedBytes := common.CopyBytes(src.Bytes())
	copiedAddress := common.BytesToAddress(copiedBytes)

	return &copiedAddress
}
