package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"sync"

	mt "github.com/cbergoon/merkletree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"

	"github.com/axiomesh/axiom-kit/types/pb"
)

// hasherPool holds LegacyKeccak256 hashers for rlpHash.
var (
	hasherPool = sync.Pool{
		New: func() any { return sha3.NewLegacyKeccak256() },
	}
)

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// Lengths of hashes and addresses in bytes.
const (
	BloomByteLength = 256
	badNibble       = ^uint64(0)
)

type Hash struct {
	rawHash common.Hash
	hashStr string
}

func NewHash(b []byte) *Hash {
	return &Hash{
		rawHash: common.BytesToHash(b),
	}
}

func decodeHash(data []byte) (*Hash, error) {
	if len(data) != common.HashLength && len(data) != 0 {
		return nil, fmt.Errorf("decode hash from bytes failed, bytes size must be %v", common.HashLength)
	}
	if len(data) == 0 {
		return &Hash{}, nil
	}

	return &Hash{
		rawHash: common.BytesToHash(data),
	}, nil
}

func NewHashByStr(s string) *Hash {
	return &Hash{
		rawHash: common.HexToHash(s),
	}
}

func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte("\"" + h.String() + "\""), nil
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	if !(len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"') {
		return errors.New("decode hash failed, invalid format")
	}
	h.rawHash = common.HexToHash(string(data[1 : len(data)-1]))
	return nil
}

// CalculateHash hashes the values of a TestContent
func (h *Hash) CalculateHash() ([]byte, error) {
	return h.Bytes(), nil
}

// Equals tests for equality of two Contents
func (h *Hash) Equals(other mt.Content) (bool, error) {
	tOther, ok := other.(*Hash)
	if !ok {
		return false, errors.New("parameter should be type TransactionHash")
	}

	return bytes.Equal(h.rawHash[:], tOther.rawHash[:]), nil
}

func (h *Hash) SetBytes(b []byte) {
	h.rawHash.SetBytes(b)
	h.hashStr = ""
}

func (h *Hash) ETHHash() common.Hash {
	return h.rawHash
}

func (h *Hash) Bytes() []byte {
	if h == nil {
		return []byte{}
	}
	return h.rawHash.Bytes()
}

func (h *Hash) String() string {
	if h.hashStr == "" {
		// if hash field is empty, initialize it for only once
		h.hashStr = h.rawHash.String()
	}

	return h.hashStr
}

func (h *Hash) Size() int {
	return common.HashLength
}

type Address struct {
	rawAddress common.Address
	addressStr string
}

func decodeAddress(data []byte) (*Address, error) {
	if len(data) != common.AddressLength && len(data) != 0 {
		return nil, fmt.Errorf("decode address from bytes failed, bytes size must be %v", common.AddressLength)
	}

	if len(data) == 0 {
		return &Address{}, nil
	}
	return &Address{
		rawAddress: common.BytesToAddress(data),
	}, nil
}

// NewAddress BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped address the left.
func NewAddress(b []byte) *Address {
	return &Address{
		rawAddress: common.BytesToAddress(b),
	}
}

func NewAddressByStr(s string) *Address {
	return &Address{
		rawAddress: common.HexToAddress(s),
	}
}

func (a *Address) MarshalJSON() ([]byte, error) {
	return []byte("\"" + a.String() + "\""), nil
}

func (a *Address) UnmarshalJSON(data []byte) error {
	if !(len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"') {
		return errors.New("decode address failed, invalid format")
	}
	a.rawAddress = common.HexToAddress(string(data[1 : len(data)-1]))
	return nil
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	a.rawAddress.SetBytes(b)
	a.addressStr = ""
}

func (a *Address) ETHAddress() common.Address {
	return a.rawAddress
}

func (a *Address) Bytes() []byte {
	if a == nil {
		return []byte{}
	}
	return a.rawAddress.Bytes()
}

// String returns an EIP55-compliant hex string representation of the address.
func (a *Address) String() string {
	if a.addressStr == "" {
		// if address field is empty, initialize it for only once
		a.addressStr = a.rawAddress.String()
	}
	return a.addressStr
}

type Bloom [BloomByteLength]byte

func decodeBloom(data []byte) (*Bloom, error) {
	if len(data) != BloomByteLength && len(data) != 0 {
		return nil, fmt.Errorf("decode bloom from bytes failed, bytes size must be %v", BloomByteLength)
	}

	b := Bloom{}
	if len(data) == 0 {
		return &b, nil
	}
	copy(b[:], data)
	return &b, nil
}

// Add adds d to the filter. Future calls of Test(d) will return true.
func (b *Bloom) Add(d []byte) {
	b.add(d, make([]byte, 6))
}

// add is internal version of Add, which takes a scratch buffer for reuse (needs to be at least 6 bytes)
func (b *Bloom) add(d []byte, buf []byte) {
	i1, v1, i2, v2, i3, v3 := bloomValues(d, buf)
	b[i1] |= v1
	b[i2] |= v2
	b[i3] |= v3
}

// Test checks if the given topic is present in the bloom filter
func (b *Bloom) Test(topic []byte) bool {
	i1, v1, i2, v2, i3, v3 := bloomValues(topic, make([]byte, 6))
	return v1 == v1&b[i1] &&
		v2 == v2&b[i2] &&
		v3 == v3&b[i3]
}

func (b *Bloom) Bytes() []byte {
	if b == nil {
		return nil
	}
	return b[:]
}

// bloomValues returns the bytes (index-value pairs) to set for the given data
func bloomValues(data []byte, hashbuf []byte) (uint, byte, uint, byte, uint, byte) {
	sha := hasherPool.Get().(KeccakState)
	sha.Reset()
	sha.Write(data)
	sha.Read(hashbuf)
	hasherPool.Put(sha)
	// The actual bits to flip
	v1 := byte(1 << (hashbuf[1] & 0x7))
	v2 := byte(1 << (hashbuf[3] & 0x7))
	v3 := byte(1 << (hashbuf[5] & 0x7))
	// The indices for the bytes to OR in
	i1 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf)&0x7ff)>>3) - 1
	i2 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[2:])&0x7ff)>>3) - 1
	i3 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[4:])&0x7ff)>>3) - 1

	return i1, v1, i2, v2, i3, v3
}

// OrBloom executes an Or operation on the bloom
func (b *Bloom) OrBloom(bl *Bloom) {
	bin := new(big.Int).SetBytes(b[:])
	input := new(big.Int).SetBytes(bl[:])
	bin.Or(bin, input)
	b.SetBytes(bin.Bytes())
}

// SetBytes sets the content of b to the given bytes.
// It panics if d is not of suitable size.
func (b *Bloom) SetBytes(d []byte) {
	if len(b) < len(d) {
		panic(fmt.Sprintf("bloom bytes too big %d %d", len(b), len(d)))
	}
	copy(b[BloomByteLength-len(d):], d)
}

func (b *Bloom) ETHBloom() types.Bloom {
	if b == nil {
		emptyBytes := [BloomByteLength]byte{}
		return types.BytesToBloom(emptyBytes[:])
	}
	return types.BytesToBloom(b[:])
}

type CodecObject interface {
	Marshal() ([]byte, error)
	Unmarshal(buf []byte) error
}

type CodecObjectConstraint[T any] interface {
	*T
	CodecObject
}

func MarshalObjects[T any, Constraint CodecObjectConstraint[T]](objs []*T) ([]byte, error) {
	objsRaw := make([][]byte, len(objs))
	for i, obj := range objs {
		objRaw, err := Constraint(obj).Marshal()
		if err != nil {
			return nil, err
		}
		objsRaw[i] = objRaw
	}
	helper := pb.BytesSlice{
		Slice: objsRaw,
	}
	return helper.MarshalVTStrict()
}

func UnmarshalObjects[T any, Constraint CodecObjectConstraint[T]](data []byte) ([]*T, error) {
	helper := pb.BytesSliceFromVTPool()
	defer helper.ReturnToVTPool()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return nil, err
	}
	objs := make([]*T, len(helper.Slice))
	for i, objRaw := range helper.Slice {
		obj := new(T)
		if err := Constraint(obj).Unmarshal(objRaw); err != nil {
			return nil, err
		}
		objs[i] = obj
	}
	return objs, nil
}

func UnmarshalObjectsWithIndex[T any, Constraint CodecObjectConstraint[T]](data []byte, index uint64) (*T, error) {
	helper := pb.BytesSliceFromVTPool()
	defer helper.ReturnToVTPool()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return nil, err
	}
	obj := new(T)
	if err := Constraint(obj).Unmarshal(helper.Slice[index]); err != nil {
		return nil, err
	}
	return obj, nil
}

func MarshalTransactions(objs []*Transaction) ([]byte, error) {
	return MarshalObjects[Transaction, *Transaction](objs)
}

func UnmarshalTransactions(data []byte) ([]*Transaction, error) {
	return UnmarshalObjects[Transaction, *Transaction](data)
}

func UnmarshalTransactionWithIndex(data []byte, index uint64) (*Transaction, error) {
	return UnmarshalObjectsWithIndex[Transaction, *Transaction](data, index)
}

func MarshalReceipts(objs []*Receipt) ([]byte, error) {
	return MarshalObjects[Receipt, *Receipt](objs)
}

func UnmarshalReceipts(data []byte) ([]*Receipt, error) {
	return UnmarshalObjects[Receipt, *Receipt](data)
}

func UnmarshalReceiptWithIndex(data []byte, index uint64) (*Receipt, error) {
	return UnmarshalObjectsWithIndex[Receipt, *Receipt](data, index)
}
