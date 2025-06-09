package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"

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

// 假设的结构体
type EpochChange struct {
	Value *int
}

// 深拷贝函数
func deepCopySlice(original []*EpochChange) []*EpochChange {
	// 创建一个新的 slice，长度与原始 slice 相同
	copied := make([]*EpochChange, len(original))

	for i, item := range original {
		if item != nil {
			// 拷贝每个元素
			newItem := &EpochChange{}

			if item.Value != nil {
				// 拷贝指针指向的值
				valueCopy := *item.Value
				newItem.Value = &valueCopy
			}

			copied[i] = newItem
		}
	}

	return copied
}

func TestName(t *testing.T) {
	val1 := 10
	val2 := 20

	slice1 := []*EpochChange{
		{Value: &val1},
		{Value: &val2},
	}

	slice2 := slice1
	go func() {
		printVal(slice2)
	}()
	slice1 = make([]*EpochChange, 0)
	time.Sleep(1 * time.Second)
}

func printVal(s []*EpochChange) {
	for {
		for _, v := range s {
			fmt.Println(*v.Value)
		}
		time.Sleep(100 * time.Millisecond)
	}

}

func Test11(t *testing.T) {
	fmt.Println(NewHash([]byte{0x0}).String())
}
