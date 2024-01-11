package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestInnerAccountSerialize(t *testing.T) {
	code := sha256.Sum256([]byte("code1"))
	ret := crypto.Keccak256Hash(code[:])
	codeHash := ret.Bytes()
	acc := &InnerAccount{
		Balance:  big.NewInt(10000000000000000),
		Nonce:    1,
		CodeHash: codeHash,
	}

	pbBlob, err := acc.Marshal()
	require.Nil(t, err)
	fmt.Printf("pbBlob size:%v\n", len(pbBlob))

	jsonBlob, err := json.Marshal(acc)
	require.Nil(t, err)
	fmt.Printf("jsonBlob size:%v\n", len(jsonBlob))

	pbAcc := &InnerAccount{}
	err = pbAcc.Unmarshal(pbBlob)
	require.Nil(t, err)
	require.False(t, pbAcc.InnerAccountChanged(acc))
	fmt.Printf("acc from pb: %v\n", pbAcc)

	jsonAcc := &InnerAccount{}
	err = json.Unmarshal(jsonBlob, jsonAcc)
	require.Nil(t, err)
	require.False(t, jsonAcc.InnerAccountChanged(acc))
	fmt.Printf("acc from json: %v\n", jsonAcc)
}
