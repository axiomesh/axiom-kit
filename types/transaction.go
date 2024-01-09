package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axiomesh/axiom-kit/types/pb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/axiomesh/axiom-kit/hexutil"
)

// deriveBufferPool holds temporary encoder buffers for DeriveSha and TX encoding.
var encodeBufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

var signer EIP155Signer = EIP155Signer{
	chainId:    big.NewInt(1),
	chainIdMul: big.NewInt(2),
}

type EIP155Signer struct {
	chainId, chainIdMul *big.Int
}

func InitEIP155Signer(chainId *big.Int) {
	if chainId == nil {
		chainId = new(big.Int)
	}
	signer = EIP155Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2)),
	}
}

func decodeSignature(sig []byte) (r, s, v *big.Int) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})

	return r, s, v
}

// signatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s EIP155Signer) signatureValues(sig []byte) (R, S, V *big.Int, err error) {
	R, S, V = decodeSignature(sig)
	if s.chainId.Sign() != 0 {
		V = big.NewInt(int64(sig[64] + 35))
		V.Add(V, s.chainIdMul)
	}
	return R, S, V, nil
}

type Eip2930Signer struct{ EIP155Signer }

// NewEIP2930Signer returns a signer that accepts EIP-2930 access list transactions,
// EIP-155 replay protected transactions, and legacy Homestead transactions.
func NewEIP2930Signer(chainId *big.Int) Eip2930Signer {
	return Eip2930Signer{EIP155Signer: EIP155Signer{chainId: chainId, chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2))}}
}

func (s Eip2930Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	txdata, ok := tx.Inner.(*AccessListTx)
	if !ok {
		return s.EIP155Signer.signatureValues(sig)
	}
	// Check that chain ID of tx matches the signer. We also accept ID zero here,
	// because it indicates that the chain ID was not specified in the tx.
	if txdata.ChainID.Sign() != 0 && txdata.ChainID.Cmp(s.chainId) != 0 {
		return nil, nil, nil, fmt.Errorf("%w: have %d want %d", types.ErrInvalidChainId, txdata.ChainID, s.chainId)
	}
	R, S, _ = decodeSignature(sig)
	V = big.NewInt(int64(sig[64]))
	return R, S, V, nil
}

type LondonSigner struct{ EIP155Signer }

func NewLondonSigner(chainId *big.Int) LondonSigner {
	return LondonSigner{EIP155Signer: EIP155Signer{chainId: chainId, chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2))}}
}

func (s LondonSigner) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	txdata, ok := tx.Inner.(*DynamicFeeTx)
	if !ok {
		return s.EIP155Signer.signatureValues(sig)
	}
	// Check that chain ID of tx matches the signer. We also accept ID zero here,
	// because it indicates that the chain ID was not specified in the tx.
	if txdata.ChainID.Sign() != 0 && txdata.ChainID.Cmp(s.chainId) != 0 {
		return nil, nil, nil, fmt.Errorf("%w: have %d want %d", types.ErrInvalidChainId, txdata.ChainID, s.chainId)
	}
	R, S, _ = decodeSignature(sig)
	V = big.NewInt(int64(sig[64]))
	return R, S, V, nil
}

// Transaction is an Ethereum transaction.
type Transaction struct {
	Inner TxData    // Consensus contents of a transaction
	Time  time.Time // Time first seen locally (spam avoidance)

	// caches
	hash atomic.Value

	size atomic.Value
	from atomic.Value
}

type writeCounter int

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

func (e *Transaction) GetVersion() []byte {
	return nil
}

func (e *Transaction) GetInner() TxData {
	return e.Inner
}

// Protected says whether the transaction is replay-protected.
func (e *Transaction) Protected() bool {
	switch tx := e.Inner.(type) {
	case *LegacyTx:
		return tx.V != nil && isProtectedV(tx.V)
	default:
		return true
	}
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28 && v != 1 && v != 0
	}
	// anything not 27 or 28 is considered protected
	return true
}

func recoverPlain(sighash *Hash, R, S, Vb *big.Int, homestead bool) (*Address, error) {
	addr, err := RecoverPlain(sighash.Bytes(), R, S, Vb, homestead)
	if err != nil {
		return nil, err
	}
	return NewAddress(addr), nil
}

func (e *Transaction) GetFrom() *Address {
	if addr := e.from.Load(); addr != nil {
		return addr.(*Address)
	}

	addr, err := e.sender()
	if err != nil {
		return nil
	}
	e.from.Store(addr)

	return addr
}

func (e *Transaction) sender() (*Address, error) {
	V, R, S := e.GetRawSignature()
	switch e.GetType() {
	case LegacyTxType:
		if !e.Protected() {
			hash := RlpHash([]any{
				e.GetNonce(),
				e.GetGasPrice(),
				e.GetGas(),
				e.Inner.GetTo(),
				e.Inner.GetValue(),
				e.GetPayload(),
			})
			addr, err := recoverPlain(NewHash(hash.Bytes()), R, S, V, true)
			if err != nil {
				return nil, errors.New("invalid signature")
			}
			return addr, nil
		}
		V = new(big.Int).Sub(V, signer.chainIdMul)
		V.Sub(V, big.NewInt(8))
	case AccessListTxType, DynamicFeeTxType, IncentiveTxType:
		// ACL txs are defined to use 0 and 1 as their recovery id, add
		// 27 to become equivalent to unprotected Homestead signatures.
		V = new(big.Int).Add(V, big.NewInt(27))
	default:
		return nil, errors.New("unknown tx type")
	}
	if e.GetChainID().Cmp(signer.chainId) != 0 {
		return nil, fmt.Errorf("invalid chain id: have %d want %d", e.GetChainID(), signer.chainId)
	}
	return recoverPlain(e.GetSignHash(), R, S, V, true)
}

func (e *Transaction) GetTo() *Address {
	if e.Inner.GetTo() == nil {
		return nil
	}
	return NewAddress(e.Inner.GetTo().Bytes())
}

func (e *Transaction) GetPayload() []byte {
	return e.Inner.GetData()
}

func (e *Transaction) GetNonce() uint64 {
	return e.Inner.GetNonce()
}

func (e *Transaction) GetValue() *big.Int {
	return e.Inner.GetValue()
}

func (e *Transaction) GetTimeStamp() int64 {
	return e.Time.Unix()
}

func (e *Transaction) GetHash() *Hash {
	if hash := e.hash.Load(); hash != nil {
		return hash.(*Hash)
	}

	var h *Hash
	if e.GetType() == LegacyTxType {
		hash := RlpHash(e.Inner)
		h = NewHash(hash.Bytes())
	} else {
		hash := PrefixedRlpHash(e.GetType(), e.Inner)
		h = NewHash(hash.Bytes())
	}
	e.hash.Store(h)
	return h
}

func (e *Transaction) GetExtra() []byte {
	return nil
}

func (e *Transaction) GetGas() uint64 {
	return e.Inner.GetGas()
}

func (e *Transaction) GetGasPrice() *big.Int {
	return e.Inner.GetGasPrice()
}

func (e *Transaction) GetGasFeeCap() *big.Int {
	return e.Inner.GetGasFeeCap()
}

func (e *Transaction) GetGasTipCap() *big.Int {
	return e.Inner.GetGasTipCap()
}

func (e *Transaction) GetChainID() *big.Int {
	return e.Inner.GetChainID()
}

func (e *Transaction) Size() int {
	if size := e.size.Load(); size != nil {
		return size.(int)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &e.Inner)
	e.size.Store(int(c))
	return int(c)
}

// Type returns the transaction type.
func (e *Transaction) GetType() byte {
	return e.Inner.TxType()
}

func (e *Transaction) GetSignature() []byte {
	var sig []byte
	v, r, s := e.Inner.RawSignatureValues()
	sig = append(sig, r.Bytes()...)
	sig = append(sig, s.Bytes()...)
	sig = append(sig, v.Bytes()...)

	return sig
}

func (e *Transaction) GetSignHash() *Hash {
	switch e.GetType() {
	case LegacyTxType:
		hash := RlpHash([]any{
			e.GetNonce(),
			e.GetGasPrice(),
			e.GetGas(),
			e.Inner.GetTo(),
			e.Inner.GetValue(),
			e.GetPayload(),
			signer.chainId, uint(0), uint(0),
		})

		return NewHash(hash.Bytes())
	case AccessListTxType:
		hash := PrefixedRlpHash(
			e.GetType(),
			[]any{
				signer.chainId,
				e.GetNonce(),
				e.GetGasPrice(),
				e.GetGas(),
				e.Inner.GetTo(),
				e.Inner.GetValue(),
				e.GetPayload(),
				e.Inner.GetAccessList(),
			})

		return NewHash(hash.Bytes())
	case DynamicFeeTxType:
		hash := PrefixedRlpHash(
			e.GetType(),
			[]any{
				signer.chainId,
				e.GetNonce(),
				e.GetGasTipCap(),
				e.GetGasFeeCap(),
				e.GetGas(),
				e.Inner.GetTo(),
				e.GetValue(),
				e.Inner.GetData(),
				e.Inner.GetAccessList(),
			})
		return NewHash(hash.Bytes())
	case IncentiveTxType:
		hash := PrefixedRlpHash(
			e.GetType(),
			[]any{
				signer.chainId,
				e.GetNonce(),
				e.GetGasTipCap(),
				e.GetGasFeeCap(),
				e.GetGas(),
				e.Inner.GetTo(),
				e.GetValue(),
				e.Inner.GetData(),
				e.Inner.GetAccessList(),
				e.Inner.GetIncentiveAddress(),
			})
		return NewHash(hash.Bytes())
	default:
		// This _should_ not happen, but in case someone sends in a bad
		// json struct via RPC, it's probably more prudent to return an
		// empty hash instead of killing the node with a panic
		// panic("Unsupported transaction type: %d", tx.typ)
		return nil
	}
}

// RawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (e *Transaction) GetRawSignature() (v, r, s *big.Int) {
	return e.Inner.RawSignatureValues()
}

func (e *Transaction) VerifySignature() error {
	if e.GetFrom() == nil {
		return errors.New("verify signature failed")
	}

	return nil
}

// // AccessList returns the access list of the transaction.
// func (e *EthTransaction) AccessList() types2.AccessList {
//	return e.Inner.GetAccessList()
// }

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	if tx.GetType() == LegacyTxType {
		return rlp.Encode(w, tx.Inner)
	}
	// It's an EIP-2718 typed TX envelope.
	buf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(buf)
	buf.Reset()
	if err := tx.encodeTyped(buf); err != nil {
		return err
	}
	return rlp.Encode(w, buf.Bytes())
}

// encodeTyped writes the canonical encoding of a typed transaction to w.
func (tx *Transaction) encodeTyped(w *bytes.Buffer) error {
	w.WriteByte(tx.GetType())
	return rlp.Encode(w, tx.Inner)
}

// MarshalBinary returns the canonical encoding of the transaction.
// For legacy transactions, it returns the RLP encoding. For EIP-2718 typed
// transactions, it returns the type and payload.
func (tx *Transaction) MarshalBinary() ([]byte, error) {
	if tx.GetType() == LegacyTxType {
		return rlp.EncodeToBytes(tx.Inner)
	}
	var buf bytes.Buffer
	err := tx.encodeTyped(&buf)
	return buf.Bytes(), err
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	kind, size, err := s.Kind()
	switch {
	case err != nil:
		return err
	case kind == rlp.List:
		// It's a legacy transaction.
		var inner LegacyTx
		err := s.Decode(&inner)
		if err == nil {
			tx.setDecoded(&inner, int(rlp.ListSize(size)))
		}
		return err
	case kind == rlp.String:
		// It's an EIP-2718 typed TX envelope.
		var b []byte
		if b, err = s.Bytes(); err != nil {
			return err
		}
		inner, err := tx.decodeTyped(b)
		if err == nil {
			tx.setDecoded(inner, len(b))
		}
		return err
	default:
		return rlp.ErrExpectedList
	}
}

// UnmarshalBinary decodes the canonical encoding of transactions.
// It supports legacy RLP transactions and EIP2718 typed transactions.
func (tx *Transaction) UnmarshalBinary(b []byte) error {
	if len(b) > 0 && b[0] > 0x7f {
		// It's a legacy transaction.
		var data LegacyTx
		err := rlp.DecodeBytes(b, &data)
		if err != nil {
			return err
		}
		tx.setDecoded(&data, len(b))
		return nil
	}
	// It's an EIP2718 typed transaction envelope.
	inner, err := tx.decodeTyped(b)
	if err != nil {
		return err
	}
	tx.setDecoded(inner, len(b))
	return nil
}

// decodeTyped decodes a typed transaction from the canonical format.
func (tx *Transaction) decodeTyped(b []byte) (TxData, error) {
	if len(b) == 0 {
		return nil, errors.New("empty tx type")
	}
	switch b[0] {
	case AccessListTxType:
		var inner AccessListTx
		err := rlp.DecodeBytes(b[1:], &inner)
		return &inner, err
	case DynamicFeeTxType:
		var inner DynamicFeeTx
		err := rlp.DecodeBytes(b[1:], &inner)
		return &inner, err
	case IncentiveTxType:
		var inner IncentiveTx
		err := rlp.DecodeBytes(b[1:], &inner)
		return &inner, err
	default:
		return nil, errors.New("unsupported tx type")
	}
}

// setDecoded sets the Inner transaction and size after decoding.
func (tx *Transaction) setDecoded(inner TxData, size int) {
	tx.Inner = inner
	tx.Time = time.Now()
	if size > 0 {
		tx.size.Store(size)
	}
}

func (e *Transaction) FromCallArgs(callArgs CallArgs) {
	if callArgs.From == nil {
		callArgs.From = &common.Address{}
	}
	e.from.Store(NewAddress(callArgs.From.Bytes()))

	inner := &AccessListTx{
		GasPrice: (*big.Int)(callArgs.GasPrice),
		To:       callArgs.To,
		Value:    (*big.Int)(callArgs.Value),
	}

	if callArgs.Gas != nil {
		inner.Gas = (uint64)(*callArgs.Gas)
	}

	if callArgs.GasPrice == nil {
		inner.GasPrice = big.NewInt(0)
	}

	if callArgs.Value == nil {
		inner.Value = big.NewInt(0)
	}

	if callArgs.Data != nil {
		inner.Data = ([]byte)(*callArgs.Data)
	}

	if callArgs.AccessList != nil {
		inner.AccessList = *callArgs.AccessList
	}

	e.Inner = inner
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	jsonM := make(map[string]any)

	jsonM["from"] = tx.GetFrom().String()
	if tx.GetTo() != nil {
		jsonM["to"] = tx.GetTo().String()
	}
	jsonM["type"] = tx.GetType()
	jsonM["Gas"] = tx.GetGas()
	jsonM["GasPrice"] = tx.GetGasPrice()
	jsonM["GasTipCap"] = tx.GetGasTipCap()
	jsonM["GasFeeCap"] = tx.GetGasFeeCap()
	jsonM["Type"] = tx.GetType()
	jsonM["Nonce"] = tx.GetNonce()
	jsonM["Value"] = tx.GetValue()
	jsonM["Hash"] = tx.GetHash()
	jsonM["ChainID"] = tx.GetChainID()
	jsonM["TimeStamp"] = tx.GetTimeStamp()
	jsonM["Signature"] = hexutil.Encode(tx.GetSignature())
	jsonM["EthTx"] = true

	return json.Marshal(jsonM)
}

func (tx *Transaction) RbftGetSize() int {
	return tx.Size()
}

func (tx *Transaction) RbftGetData() []byte {
	return tx.Inner.GetData()
}

func (tx *Transaction) RbftIsConfigTx() bool {
	return false
}

func (tx *Transaction) RbftGetTxHash() string {
	return tx.GetHash().String()
}

func (tx *Transaction) RbftGetFrom() string {
	return tx.GetFrom().String()
}

func (tx *Transaction) RbftGetTimeStamp() int64 {
	return tx.Time.UnixNano()
}

func (tx *Transaction) RbftGetNonce() uint64 {
	return tx.GetNonce()
}

func (tx *Transaction) RbftUnmarshal(raw []byte) error {
	pbTx := &pb.Transaction{}
	if err := pbTx.UnmarshalVT(raw); err != nil {
		return err
	}
	tx.fromPB(pbTx)
	return nil
}

func (tx *Transaction) RbftMarshal() ([]byte, error) {
	pbTx := tx.toPB()
	return pbTx.MarshalVT()
}

func (tx *Transaction) Unmarshal(buf []byte) error {
	return tx.UnmarshalBinary(buf)
}

func (tx *Transaction) Marshal() ([]byte, error) {
	return tx.MarshalBinary()
}

func (tx *Transaction) SignByTxType(prv *ecdsa.PrivateKey) error {
	switch tx.GetType() {
	case LegacyTxType:
		return tx.Sign(prv)
	case AccessListTxType:
		acSigner := NewEIP2930Signer(signer.chainId)

		h := PrefixedRlpHash(
			tx.GetType(),
			[]any{
				signer.chainId,
				tx.GetNonce(),
				tx.Inner.GetGasPrice(),
				tx.Inner.GetGas(),
				tx.Inner.GetTo(),
				tx.Inner.GetValue(),
				tx.Inner.GetData(),
				tx.Inner.GetAccessList(),
			})
		sig, err := crypto.Sign(h.Bytes(), prv)
		if err != nil {
			return err
		}
		r, s, v, err := acSigner.SignatureValues(tx, sig)
		if err != nil {
			return err
		}
		tx.Inner.setSignatureValues(signer.chainId, v, r, s)
		return nil

	case DynamicFeeTxType:
		dynamicSigner := NewLondonSigner(signer.chainId)
		h := PrefixedRlpHash(
			tx.GetType(),
			[]any{
				signer.chainId,
				tx.GetNonce(),
				tx.Inner.GetGasTipCap(),
				tx.Inner.GetGasFeeCap(),
				tx.Inner.GetGas(),
				tx.Inner.GetTo(),
				tx.Inner.GetValue(),
				tx.Inner.GetData(),
				tx.Inner.GetAccessList(),
			})
		sig, err := crypto.Sign(h.Bytes(), prv)
		if err != nil {
			return err
		}

		r, s, v, err := dynamicSigner.SignatureValues(tx, sig)
		if err != nil {
			return err
		}
		tx.Inner.setSignatureValues(signer.chainId, v, r, s)
	case IncentiveTxType:
		incentiveSigner := NewLondonSigner(signer.chainId)
		h := PrefixedRlpHash(
			tx.GetType(),
			[]any{
				signer.chainId,
				tx.GetNonce(),
				tx.Inner.GetGasTipCap(),
				tx.Inner.GetGasFeeCap(),
				tx.Inner.GetGas(),
				tx.Inner.GetTo(),
				tx.Inner.GetValue(),
				tx.Inner.GetData(),
				tx.Inner.GetAccessList(),
				tx.Inner.GetIncentiveAddress(),
			})
		sig, err := crypto.Sign(h.Bytes(), prv)
		if err != nil {
			return err
		}

		r, s, v, err := incentiveSigner.SignatureValues(tx, sig)
		if err != nil {
			return err
		}
		tx.Inner.setSignatureValues(signer.chainId, v, r, s)

	}
	return nil
}

func (tx *Transaction) Sign(prv *ecdsa.PrivateKey) error {
	h := RlpHash([]any{
		tx.GetInner().GetNonce(),
		tx.GetInner().GetGasPrice(),
		tx.GetInner().GetGas(),
		tx.GetInner().GetTo(),
		tx.GetInner().GetValue(),
		tx.GetInner().GetData(),
		signer.chainId, uint(0), uint(0),
	})
	sig, err := crypto.Sign(h.Bytes(), prv)
	if err != nil {
		return err
	}

	r, s, v, err := signer.signatureValues(sig)
	if err != nil {
		return err
	}
	tx.Inner.setSignatureValues(signer.chainId, v, r, s)
	return nil
}

func (tx *Transaction) toPB() *pb.Transaction {
	if tx == nil || tx.Inner == nil {
		return nil
	}

	var txData pb.TxDataVariant
	switch tx := tx.Inner.(type) {
	case *AccessListTx:
		txData.TxDataType = &pb.TxDataVariant_AccessListTx{
			AccessListTx: tx.toPB(),
		}
	case *LegacyTx:
		txData.TxDataType = &pb.TxDataVariant_LegacyTx{
			LegacyTx: tx.toPB(),
		}
	case *DynamicFeeTx:
		txData.TxDataType = &pb.TxDataVariant_DynamicFeeTx{
			DynamicFeeTx: tx.toPB(),
		}
	case *IncentiveTx:
		txData.TxDataType = &pb.TxDataVariant_IncentiveTx{
			IncentiveTx: tx.toPB(),
		}
	}

	return &pb.Transaction{
		Inner: &txData,
		Time:  toPBTime(tx.Time),
	}
}

func (tx *Transaction) fromPB(PBTx *pb.Transaction) {
	if PBTx == nil {
		return
	}

	switch x := PBTx.Inner.TxDataType.(type) {
	case *pb.TxDataVariant_AccessListTx:
		txData := &AccessListTx{}
		txData.fromPB(x.AccessListTx)
		tx.Inner = txData
	case *pb.TxDataVariant_LegacyTx:
		txData := &LegacyTx{}
		txData.fromPB(x.LegacyTx)
		tx.Inner = txData
	case *pb.TxDataVariant_DynamicFeeTx:
		txData := &DynamicFeeTx{}
		txData.fromPB(x.DynamicFeeTx)
		tx.Inner = txData
	case *pb.TxDataVariant_IncentiveTx:
		TxData := &IncentiveTx{}
		TxData.fromPB(x.IncentiveTx)
		tx.Inner = TxData
	}
	// ignore time
}

func toPBTime(t time.Time) int64 {
	return t.UnixNano()
}

func fromPBTime(timestamp int64) time.Time {
	return time.Unix(0, timestamp)
}

type Signer struct {
	Sk   *ecdsa.PrivateKey
	Addr *Address
}

func LoadSignerWithPk(pk string) (*Signer, error) {
	privateKeyHex := strings.TrimPrefix(pk, "0x")
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key error: %v", err)
	}

	privateKey := &ecdsa.PrivateKey{}
	privateKey.Curve = secp256k1.S256()
	privateKey.D = new(big.Int).SetBytes(privateKeyBytes)

	privateKey.X, privateKey.Y = secp256k1.S256().ScalarBaseMult(privateKey.D.Bytes())

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	s := &Signer{
		Sk:   privateKey,
		Addr: NewAddress(addr.Bytes()),
	}
	return s, nil
}

func GenerateSigner() (*Signer, error) {
	sk, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	a := crypto.PubkeyToAddress(sk.PublicKey)

	return &Signer{
		Sk:   sk,
		Addr: NewAddress(a.Bytes()),
	}, nil
}

func GenerateTransactionWithSigner(nonce uint64, to *Address, value *big.Int, data []byte, s *Signer) (*Transaction, error) {
	if data == nil {
		data = []byte{}
	}
	inner := &LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(5),
		Gas:      30000000,
		Value:    value,
		Data:     data,
	}
	if to != nil {
		t := to.ETHAddress()
		inner.To = &t
	}
	tx := &Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.Sign(s.Sk); err != nil {
		return nil, err
	}
	return tx, nil
}

func GenerateTransactionAndSigner(nonce uint64, to *Address, value *big.Int, data []byte) (*Transaction, *Signer, error) {
	s, err := GenerateSigner()
	if err != nil {
		return nil, nil, err
	}
	tx, err := GenerateTransactionWithSigner(nonce, to, value, data, s)
	if err != nil {
		return nil, nil, err
	}

	return tx, s, nil
}

func GenerateEmptyTransactionAndSigner() (*Transaction, error) {
	s, err := GenerateSigner()
	if err != nil {
		return nil, err
	}
	tx, err := GenerateTransactionWithSigner(0, NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(0), nil, s)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// NOTICE!! this function just supoort for unit test case
func GenerateWrongSignTransactionAndSigner(illegalSign bool) (*Transaction, *Signer, error) {
	s, err := GenerateSigner()
	if err != nil {
		return nil, nil, err
	}
	tx, err := GenerateTransactionWithSigner(0, NewAddressByStr("0xdAC17F958D2ee523a2206206994597C13D831ec7"), big.NewInt(0), []byte("hello"), s)
	if err != nil {
		return nil, nil, err
	}

	if illegalSign {
		tx.Inner.(*LegacyTx).S = big.NewInt(0)
	} else {
		tx.Inner.(*LegacyTx).Data = []byte("world")
	}
	return tx, s, nil
}

func GenerateAccessListTxAndSigner() (*Transaction, error) {
	s, err := GenerateSigner()
	if err != nil {
		return nil, err
	}
	to := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	inner := &AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    1,
		GasPrice: big.NewInt(500),
		Gas:      0x5f5e100,
		To:       &to,
		Value:    big.NewInt(0),
		AccessList: types.AccessList{
			types.AccessTuple{
				Address:     common.HexToAddress("0x01"),
				StorageKeys: []common.Hash{common.HexToHash("0x01")},
			},
		},
	}
	tx := &Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.SignByTxType(s.Sk); err != nil {
		return nil, err
	}
	return tx, nil
}

func GenerateDynamicFeeTxAndSinger() (*Transaction, error) {
	s, err := GenerateSigner()
	if err != nil {
		return nil, err
	}
	to := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	inner := &DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     0,
		GasTipCap: big.NewInt(8000000000000),
		GasFeeCap: big.NewInt(8000000000000),
		Gas:       0x5f5e100,
		To:        &to,
		Data:      []byte{},
		Value:     big.NewInt(0),
	}
	tx := &Transaction{
		Inner: inner,
		Time:  time.Now(),
	}

	if err := tx.SignByTxType(s.Sk); err != nil {
		return nil, err
	}
	return tx, nil
}

func toPBBigInt(bi *big.Int) *pb.BigInt {
	if bi == nil {
		return nil
	}
	return &pb.BigInt{Value: bi.String()}
}

func fromPBBigInt(bi *pb.BigInt) *big.Int {
	if bi == nil {
		return nil
	}
	n := new(big.Int)
	n.SetString(bi.Value, 10)
	return n
}

func toPBAddress(addr *common.Address) []byte {
	if addr == nil {
		return nil
	}
	return addr.Bytes()
}

func fromPBAddress(pbAddr []byte) *common.Address {
	if pbAddr == nil {
		return nil
	}
	var addr common.Address
	addr.SetBytes(pbAddr)
	return &addr
}

func toPBAccessList(ac types.AccessList) *pb.AccessList {
	if ac == nil {
		return nil
	}
	accessListPB := make([]*pb.AccessTuple, len(ac))
	for i, tuple := range ac {
		accessListPB[i] = &pb.AccessTuple{
			Address:     tuple.Address[:],
			StorageKeys: toPBHashes(tuple.StorageKeys),
		}
	}
	AccessList := &pb.AccessList{
		AccessTuples: accessListPB,
	}
	return AccessList
}

func fromPBAccessList(accessListPB *pb.AccessList) types.AccessList {
	if accessListPB == nil {
		return nil
	}
	accessList := make(types.AccessList, len(accessListPB.AccessTuples))
	for i, tuplePB := range accessListPB.AccessTuples {
		accessList[i] = types.AccessTuple{
			Address:     common.BytesToAddress(tuplePB.Address),
			StorageKeys: fromPBHashes(tuplePB.StorageKeys),
		}
	}
	return accessList
}

func fromPBHashes(hashesPB [][]byte) []common.Hash {
	if hashesPB == nil {
		return nil
	}
	hashes := make([]common.Hash, len(hashesPB))
	for i, hashPB := range hashesPB {
		hashes[i] = common.BytesToHash(hashPB)
	}
	return hashes
}

func toPBHashes(hashes []common.Hash) [][]byte {
	if hashes == nil {
		return nil
	}
	hashesPB := make([][]byte, len(hashes))
	for i, hash := range hashes {
		hashesPB[i] = hash.Bytes()
	}
	return hashesPB
}
