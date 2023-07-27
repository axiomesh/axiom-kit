package types

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	hash0   = "0x9f41dd84524bf8a42f8ab58ecfca6e1752d6fd93fe8dc00af4c71963c97db59f"
	account = "0x929545f44692178EDb7FA468B44c5351596184Ba"
)

func TestHash(t *testing.T) {
	hash1 := NewHashByStr(hash0)
	require.Equal(t, hash0, hash1.String())
	require.Equal(t, hash0, hash1.String())

	// test address SetBytes and Set
	hash2 := &Hash{}
	hash2.SetBytes(hash1.Bytes())
	require.Equal(t, true, bytes.Equal(hash1.Bytes(), hash2.Bytes()))
	require.Equal(t, hash1.String(), hash2.String())
	require.Equal(t, hash1.String(), hash2.String())

	bytes, err := json.Marshal(hash1)
	assert.Nil(t, err)
	decodeHash := &Hash{}
	err = json.Unmarshal(bytes, decodeHash)
	assert.Nil(t, err)
	assert.Equal(t, hash1.Bytes(), decodeHash.Bytes())
}

func TestAddress(t *testing.T) {
	addr1 := NewAddressByStr(account)

	require.Equal(t, account, addr1.String())
	require.Equal(t, account, addr1.String())

	// test address SetBytes and Set
	addr2 := &Address{}
	addr2.SetBytes(addr1.Bytes())
	require.Equal(t, true, bytes.Equal(addr1.Bytes(), addr2.Bytes()))
	require.Equal(t, addr1.String(), addr2.String())
	require.Equal(t, addr1.String(), addr2.String())

	bytes, err := json.Marshal(addr1)
	assert.Nil(t, err)
	decodeAddress := &Address{}
	err = json.Unmarshal(bytes, decodeAddress)
	assert.Nil(t, err)
	assert.Equal(t, addr1.Bytes(), decodeAddress.Bytes())
}

func TestBloom(t *testing.T) {
	var bin Bloom
	bin.Add([]byte(hash0))
	ok := bin.Test([]byte(hash0))
	require.True(t, ok)

	formalHash := "0x9f41dd84524bf8a42f8ab58ecfca6e1752d6fd93fe8dc00af4c71963c97db59k"
	var bin1 Bloom
	bin1.Add([]byte(formalHash))
	ok = bin1.Test([]byte(formalHash))
	require.True(t, ok)

	bin.OrBloom(&bin1)
	ok = bin.Test([]byte(formalHash))
	require.True(t, ok)

}
