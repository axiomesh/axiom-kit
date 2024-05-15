package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type NoZeroInitializer interface {
	InitNoZero()
}

func InitializeValue(v any) error {
	value := reflect.ValueOf(v).Elem()
	return initializeValue(value)
}

func initializeValue(value reflect.Value) error {
	valueType := value.Type()
	switch value.Kind() {
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value.SetUint(1)
	case reflect.Float32, reflect.Float64:
		value.SetFloat(1.1)
	case reflect.Array:
		if valueType.Len() != 0 {
			elemType := valueType.Elem()
			elem := reflect.New(elemType).Elem()
			if err := initializeValue(elem); err != nil {
				return err
			}
			value.Index(0).Set(elem)
		}
	case reflect.Map:
		if value.IsNil() {
			value.Set(reflect.MakeMap(valueType))
			keyType := valueType.Key()
			valType := valueType.Elem()
			key := reflect.New(keyType).Elem()
			val := reflect.New(valType).Elem()
			if err := initializeValue(key); err != nil {
				return err
			}
			if err := initializeValue(val); err != nil {
				return err
			}
			value.SetMapIndex(key, val)
		}
	case reflect.Pointer:
		if value.IsNil() {
			value.Set(reflect.New(valueType.Elem()))
		}
		noZeroInitializer, ok := value.Interface().(NoZeroInitializer)
		if !ok {
			return errors.Errorf("pointer type must implement NoZeroInitializer interface: %v", value.Type())
		}
		noZeroInitializer.InitNoZero()
	case reflect.Slice:
		value.Set(reflect.MakeSlice(valueType, 1, 1))
		elemType := valueType.Elem()
		elem := reflect.New(elemType).Elem()
		if err := initializeValue(elem); err != nil {
			return err
		}
		value.Index(0).Set(elem)
	case reflect.String:
		value.SetString("1")
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if err := initializeValue(value.Field(i)); err != nil {
				return err
			}
		}
	default:
		return errors.Errorf("unsupported type: %v for initialization", value.Type())
	}
	return nil
}

type InnerStruct2 struct {
	InnerField int
}

func (i *InnerStruct2) InitNoZero() {
	i.InnerField = 1
}

func TestInitializeValue(t *testing.T) {
	type InnerStruct struct {
		InnerField int
	}

	type TestStruct struct {
		StringField    string
		IntField       int
		SliceField     []InnerStruct
		BytesField     []byte
		ArrayField     [1]InnerStruct
		ByteArrayField [1]byte
		BoolField      bool
		FloatField     float64
		MapField       map[string]InnerStruct
		PtrField       *InnerStruct
		StructFiled    InnerStruct
	}

	v1 := &TestStruct{}
	err := InitializeValue(v1)
	require.ErrorContains(t, err, "pointer type must implement NoZeroInitializer interface")

	type TestStruct2 struct {
		StringField    string
		IntField       int
		SliceField     []InnerStruct2
		BytesField     []byte
		ArrayField     [1]InnerStruct2
		ByteArrayField [1]byte
		BoolField      bool
		FloatField     float64
		MapField       map[string]InnerStruct2
		PtrField       *InnerStruct2
		StructFiled    InnerStruct2
	}
	v2 := &TestStruct2{}
	require.Nil(t, InitializeValue(v2))

	res, _ := json.Marshal(v2)
	fmt.Println(string(res))
}

func TestEpochInfo_Clone(t *testing.T) {
	epochInfo := &EpochInfo{}
	require.Nil(t, InitializeValue(epochInfo))
	require.True(t, reflect.DeepEqual(epochInfo, epochInfo.Clone()))
}

func getAvailableEpochInfoWithSetter(setter func(e *EpochInfo)) *EpochInfo {
	e := &EpochInfo{
		Epoch:       1,
		EpochPeriod: 100,
		StartBlock:  1,
		ConsensusParams: ConsensusParams{
			ProposerElectionType:          ProposerElectionTypeWRF,
			CheckpointPeriod:              1,
			HighWatermarkCheckpointPeriod: 10,
			MinValidatorNum:               4,
			MaxValidatorNum:               4,
			BlockMaxTxNum:                 500,
			EnableTimedGenEmptyBlock:      false,
			NotActiveWeight:               1,
			AbnormalNodeExcludeView:       10,
			AgainProposeIntervalBlockInValidatorsNumPercentage: 30,
			ContinuousNullRequestToleranceNumber:               3,
			ReBroadcastToleranceNumber:                         2,
		},
		FinanceParams: FinanceParams{
			GasLimit:    0x5f5e100,
			MinGasPrice: CoinNumberByGmol(1000),
		},
		StakeParams: StakeParams{
			StakeEnable:                      true,
			MaxAddStakeRatio:                 1000,
			MaxUnlockStakeRatio:              1000,
			UnlockPeriod:                     10,
			MaxPendingInactiveValidatorRatio: 10,
			MinDelegateStake:                 CoinNumberByAxc(100),
			MinValidatorStake:                CoinNumberByAxc(10000000),
			MaxValidatorStake:                CoinNumberByAxc(50000000),
		},
		MiscParams: MiscParams{
			TxMaxSize: 4 * 32 * 1024,
		},
	}
	if setter != nil {
		setter(e)
	}

	return e
}

func TestEpochInfo_Validate(t *testing.T) {
	tests := []struct {
		name       string
		epochInfo  *EpochInfo
		checkError func(t *testing.T, err error)
	}{
		{
			name:      "valid",
			epochInfo: getAvailableEpochInfoWithSetter(nil),
			checkError: func(t *testing.T, err error) {
				require.Nil(t, err)
			},
		},
		{
			name: "err EpochPeriod zero",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.EpochPeriod = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "epoch_period cannot be 0")
			},
		},
		{
			name: "err EpochPeriod",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.CheckpointPeriod = 10
				e.EpochPeriod = 15
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "epoch_period must be an integral multiple of checkpoint_period")
			},
		},

		// consensus params
		{
			name: "err CheckpointPeriod",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.CheckpointPeriod = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "checkpoint_period cannot be 0")
			},
		},
		{
			name: "err HighWatermarkCheckpointPeriod",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.HighWatermarkCheckpointPeriod = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "high_watermark_checkpoint_period cannot be 0")
			},
		},
		{
			name: "err AgainProposeIntervalBlockInValidatorsNumPercentage zero",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "again_propose_interval_block_in_validators_num_percentage cannot be 0")
			},
		},
		{
			name: "err AgainProposeIntervalBlockInValidatorsNumPercentage",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage = 1000
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "again_propose_interval_block_in_validators_num_percentage cannot be greater than or equal to 100")
			},
		},
		{
			name: "err MinValidatorNum",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.MinValidatorNum = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "min_validator_num must be greater than 0")
			},
		},
		{
			name: "err MaxValidatorNum",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.MaxValidatorNum = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_validator_num must be greater than or equal to 4")
			},
		},
		{
			name: "err BlockMaxTxNum",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.BlockMaxTxNum = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "block_max_tx_num cannot be 0")
			},
		},
		{
			name: "err AbnormalNodeExcludeView",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.AbnormalNodeExcludeView = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "exclude_view cannot be 0")
			},
		},
		{
			name: "err ProposerElectionType",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.ConsensusParams.ProposerElectionType = "sss"
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "unsupported proposer_election_type")
			},
		},

		// finance params
		{
			name: "err MinGasPrice",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.FinanceParams.MinGasPrice = CoinNumberByAxc(1000000)
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "invalid min_gas_price")
			},
		},

		// StakeParams
		{
			name: "err MaxAddStakeRatio zero",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxAddStakeRatio = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_add_stake_ratio cannot be 0")
			},
		},
		{
			name: "err MaxAddStakeRatio overflow",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxAddStakeRatio = RatioLimit + 1
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_add_stake_ratio cannot be greater than")
			},
		},
		{
			name: "err MaxUnlockStakeRatio zero",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxUnlockStakeRatio = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_unlock_stake_ratio cannot be 0")
			},
		},
		{
			name: "err MaxUnlockStakeRatio overflow",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxUnlockStakeRatio = RatioLimit + 1
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_unlock_stake_ratio cannot be greater than")
			},
		},
		{
			name: "err MaxPendingInactiveValidatorRatio zero",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxPendingInactiveValidatorRatio = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_pending_inactive_validator_ratio cannot be 0")
			},
		},
		{
			name: "err MaxPendingInactiveValidatorRatio overflow",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxPendingInactiveValidatorRatio = RatioLimit + 1
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "max_pending_inactive_validator_ratio cannot be greater than")
			},
		},
		{
			name: "err MinDelegateStake",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MinDelegateStake = CoinNumberByBigInt(new(big.Int).SetInt64(-1))
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "invalid min_delegate_stake")
			},
		},
		{
			name: "err MinValidatorStake",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MinValidatorStake = CoinNumberByBigInt(new(big.Int).SetInt64(-1))
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "invalid min_validator_stake")
			},
		},
		{
			name: "err MaxValidatorStake",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.StakeParams.MaxValidatorStake = CoinNumberByBigInt(new(big.Int).SetInt64(-1))
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "invalid max_validator_stake")
			},
		},

		// misc params
		{
			name: "err TxMaxSize",
			epochInfo: getAvailableEpochInfoWithSetter(func(e *EpochInfo) {
				e.MiscParams.TxMaxSize = 0
			}),
			checkError: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "tx_max_size cannot be 0")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkError(t, tt.epochInfo.Validate())
		})
	}
}

func TestEpochInfo_Marshal(t *testing.T) {
	epochInfo := &EpochInfo{}
	require.Nil(t, InitializeValue(epochInfo))

	raw, err := epochInfo.Marshal()
	require.Nil(t, err)
	epochInfo2 := &EpochInfo{}
	err = epochInfo2.Unmarshal(raw)
	require.Nil(t, err)
	require.True(t, reflect.DeepEqual(epochInfo, epochInfo2))
}
