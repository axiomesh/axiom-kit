package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeInfo_Marshal(t *testing.T) {
	nodeInfo := &NodeInfo{}
	require.Nil(t, InitializeValue(nodeInfo))

	raw, err := nodeInfo.Marshal()
	require.Nil(t, err)
	nodeInfo2 := &NodeInfo{}
	err = nodeInfo2.Unmarshal(raw)
	require.Nil(t, err)
	require.True(t, reflect.DeepEqual(nodeInfo, nodeInfo2))
}
