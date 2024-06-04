package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
)

const (
	ProposerElectionTypeWRF              = "wrf"
	ProposerElectionTypeAbnormalRotation = "abnormal-rotation"

	RatioLimit = 10000
)

type ConsensusParams struct {
	// The proposer election type, default is wrf
	// wrf: WRF
	// abnormal-rotation: rotating by view(pbft logic, disable auto change proposer)
	ProposerElectionType string `mapstructure:"proposer_election_type" toml:"proposer_election_type" json:"proposer_election_type"`

	// The number of sustained blocks per Checkpoint.
	CheckpointPeriod uint64 `mapstructure:"checkpoint_period" toml:"checkpoint_period" json:"checkpoint_period"`

	// Used to calculate max log size in memory: CheckpointPeriod*HighWatermarkCheckpointPeriod.
	HighWatermarkCheckpointPeriod uint64 `mapstructure:"high_watermark_checkpoint_period" toml:"high_watermark_checkpoint_period" json:"high_watermark_checkpoint_period"`

	// The minimum number of validators in the network.
	MinValidatorNum uint64 `mapstructure:"min_validator_num" toml:"min_validator_num" json:"min_validator_num"`

	// The maximum number of validators in the network.
	MaxValidatorNum uint64 `mapstructure:"max_validator_num" toml:"max_validator_num" json:"max_validator_num"`

	// The maximum number of packaged transactions per block.
	BlockMaxTxNum uint64 `mapstructure:"block_max_tx_num" toml:"block_max_tx_num" json:"block_max_tx_num"`

	// Enable timed gen empty block feature.
	EnableTimedGenEmptyBlock bool `mapstructure:"enable_timed_gen_empty_block" toml:"enable_timed_gen_empty_block" json:"enable_timed_gen_empty_block"`

	// The weight of the faulty node after viewchange is triggered is set to the weight so that the node has a low probability of blocking.
	NotActiveWeight int64 `mapstructure:"not_active_weight" toml:"not_active_weight" json:"not_active_weight"`

	// The low weight of the viewchange node is restored to normal after the specified number of rounds.
	AbnormalNodeExcludeView uint64 `mapstructure:"abnormal_node_exclude_view" toml:"abnormal_node_exclude_view" json:"abnormal_node_exclude_view"`

	// The block interval for node to propose again in validators num percentage,
	// Ensure that a node cannot continuously produce blocks
	// min is 1, max is validatorSetNum - 1
	AgainProposeIntervalBlockInValidatorsNumPercentage uint64 `mapstructure:"again_propose_interval_block_in_validators_num_percentage" toml:"again_propose_interval_block_in_validators_num_percentage" json:"again_propose_interval_block_in_validators_num_percentage"`

	// ContinuousNullRequestToleranceNumber Viewchange will be sent when there is a packageable transaction locally and n nullrequests are received consecutively.
	ContinuousNullRequestToleranceNumber uint64 `mapstructure:"continuous_null_request_tolerance_number" toml:"continuous_null_request_tolerance_number" json:"continuous_null_request_tolerance_number"`

	// ReBroadcastToleranceNumber replicate will rebroadcast pending ready txs when receiving null requests from primary above the threshold
	// !!! notice!!! this param must smaller than  ContinuousNullRequestToleranceNumber
	ReBroadcastToleranceNumber uint64 `mapstructure:"rebroadcast_tolerance_number" toml:"rebroadcast_tolerance_number" json:"rebroadcast_tolerance_number"`
}

type FinanceParams struct {
	GasLimit uint64 `mapstructure:"gas_limit" toml:"gas_limit" json:"gas_limit"`

	MinGasPrice *CoinNumber `mapstructure:"min_gas_price" toml:"min_gas_price" json:"min_gas_price"`
}

type StakeParams struct {
	StakeEnable           bool   `mapstructure:"stake_enable" toml:"stake_enable" json:"stake_enable"`
	MaxAddStakeRatio      uint64 `mapstructure:"max_add_stake_ratio" toml:"max_add_stake_ratio" json:"max_add_stake_ratio"`
	MaxUnlockStakeRatio   uint64 `mapstructure:"max_unlock_stake_ratio" toml:"max_unlock_stake_ratio" json:"max_unlock_stake_ratio"`
	MaxUnlockingRecordNum uint64 `mapstructure:"max_unlocking_record_num" toml:"max_unlocking_record_num" json:"max_unlocking_record_num"`

	// unit: seconds
	UnlockPeriod uint64 `mapstructure:"unlock_period" toml:"unlock_period" json:"unlock_period"`

	MaxPendingInactiveValidatorRatio uint64      `mapstructure:"max_pending_inactive_validator_ratio" toml:"max_pending_inactive_validator_ratio" json:"max_pending_inactive_validator_ratio"`
	MinDelegateStake                 *CoinNumber `mapstructure:"min_delegate_stake" toml:"min_delegate_stake" json:"min_delegate_stake"`
	MinValidatorStake                *CoinNumber `mapstructure:"min_validator_stake" toml:"min_validator_stake" json:"min_validator_stake"`
	MaxValidatorStake                *CoinNumber `mapstructure:"max_validator_stake" toml:"max_validator_stake" json:"max_validator_stake"`
	EnablePartialUnlock              bool        `mapstructure:"enable_partial_unlock" toml:"enable_partial_unlock" json:"enable_partial_unlock"`
}

type MiscParams struct {
	TxMaxSize uint64 `mapstructure:"tx_max_size" toml:"tx_max_size" json:"tx_max_size"`
}

type EpochInfo struct {
	// Epoch number.
	Epoch uint64 `mapstructure:"epoch" toml:"epoch" json:"epoch"`

	// The number of blocks lasting per Epoch (must be a multiple of the CheckpointPeriod).
	EpochPeriod uint64 `mapstructure:"epoch_period" toml:"epoch_period" json:"epoch_period"`

	// Epoch start block.
	StartBlock uint64 `mapstructure:"start_block" toml:"start_block" json:"start_block"`

	// Consensus params.
	ConsensusParams ConsensusParams `mapstructure:"consensus_params" toml:"consensus_params" json:"consensus_params"`

	// FinanceParams params about gas
	FinanceParams FinanceParams `mapstructure:"finance_params" toml:"finance_params" json:"finance_params"`

	// StakeParams params about stake
	StakeParams StakeParams `mapstructure:"stake_params" toml:"stake_params" json:"stake_params"`

	MiscParams MiscParams `mapstructure:"misc_params" toml:"misc_params" json:"misc_params"`
}

func (e *EpochInfo) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *EpochInfo) Unmarshal(raw []byte) error {
	return json.Unmarshal(raw, e)
}

func (e *EpochInfo) Clone() *EpochInfo {
	return &EpochInfo{
		Epoch:       e.Epoch,
		EpochPeriod: e.EpochPeriod,
		StartBlock:  e.StartBlock,
		ConsensusParams: ConsensusParams{
			ProposerElectionType:          e.ConsensusParams.ProposerElectionType,
			CheckpointPeriod:              e.ConsensusParams.CheckpointPeriod,
			HighWatermarkCheckpointPeriod: e.ConsensusParams.HighWatermarkCheckpointPeriod,
			MinValidatorNum:               e.ConsensusParams.MinValidatorNum,
			MaxValidatorNum:               e.ConsensusParams.MaxValidatorNum,
			BlockMaxTxNum:                 e.ConsensusParams.BlockMaxTxNum,
			EnableTimedGenEmptyBlock:      e.ConsensusParams.EnableTimedGenEmptyBlock,
			NotActiveWeight:               e.ConsensusParams.NotActiveWeight,
			AbnormalNodeExcludeView:       e.ConsensusParams.AbnormalNodeExcludeView,
			AgainProposeIntervalBlockInValidatorsNumPercentage: e.ConsensusParams.AgainProposeIntervalBlockInValidatorsNumPercentage,
			ContinuousNullRequestToleranceNumber:               e.ConsensusParams.ContinuousNullRequestToleranceNumber,
			ReBroadcastToleranceNumber:                         e.ConsensusParams.ReBroadcastToleranceNumber,
		},
		FinanceParams: FinanceParams{
			GasLimit:    e.FinanceParams.GasLimit,
			MinGasPrice: e.FinanceParams.MinGasPrice.Clone(),
		},
		StakeParams: StakeParams{
			StakeEnable:                      e.StakeParams.StakeEnable,
			MaxAddStakeRatio:                 e.StakeParams.MaxAddStakeRatio,
			MaxUnlockStakeRatio:              e.StakeParams.MaxUnlockStakeRatio,
			MaxUnlockingRecordNum:            e.StakeParams.MaxUnlockingRecordNum,
			UnlockPeriod:                     e.StakeParams.UnlockPeriod,
			MaxPendingInactiveValidatorRatio: e.StakeParams.MaxPendingInactiveValidatorRatio,
			MinDelegateStake:                 e.StakeParams.MinDelegateStake.Clone(),
			MinValidatorStake:                e.StakeParams.MinValidatorStake.Clone(),
			MaxValidatorStake:                e.StakeParams.MaxValidatorStake.Clone(),
			EnablePartialUnlock:              e.StakeParams.EnablePartialUnlock,
		},
		MiscParams: MiscParams{
			TxMaxSize: e.MiscParams.TxMaxSize,
		},
	}
}

func (e *EpochInfo) Validate() error {
	if err := func() error {
		if err := e.ConsensusParams.Validate(); err != nil {
			return errors.Wrap(err, "consensus params validate failed")
		}
		if err := e.FinanceParams.Validate(); err != nil {
			return errors.Wrap(err, "finance params validate failed")
		}
		if err := e.StakeParams.Validate(); err != nil {
			return errors.Wrap(err, "stake params validate failed")
		}
		if err := e.MiscParams.Validate(); err != nil {
			return errors.Wrap(err, "misc params validate failed")
		}

		if e.EpochPeriod == 0 {
			return errors.New("epoch_period cannot be 0")
		} else if e.EpochPeriod%e.ConsensusParams.CheckpointPeriod != 0 {
			return errors.New("epoch_period must be an integral multiple of checkpoint_period")
		}

		return nil
	}(); err != nil {
		return errors.Wrap(err, "epoch info validate failed")
	}

	return nil
}

func (p *ConsensusParams) Validate() error {
	if p.CheckpointPeriod == 0 {
		return errors.New("checkpoint_period cannot be 0")
	}

	if p.HighWatermarkCheckpointPeriod == 0 {
		return errors.New("high_watermark_checkpoint_period cannot be 0")
	}

	if p.AgainProposeIntervalBlockInValidatorsNumPercentage == 0 {
		return errors.New("again_propose_interval_block_in_validators_num_percentage cannot be 0")
	} else if p.AgainProposeIntervalBlockInValidatorsNumPercentage >= 100 {
		return errors.New("again_propose_interval_block_in_validators_num_percentage cannot be greater than or equal to 100")
	}

	if p.MinValidatorNum == 0 {
		return errors.New("min_validator_num must be greater than 0")
	}

	if p.MaxValidatorNum < p.MinValidatorNum {
		return errors.Errorf("max_validator_num must be greater than or equal to %d", p.MinValidatorNum)
	}

	if p.BlockMaxTxNum == 0 {
		return errors.New("block_max_tx_num cannot be 0")
	}

	if p.AbnormalNodeExcludeView == 0 {
		return errors.New("exclude_view cannot be 0")
	}

	if p.ProposerElectionType != ProposerElectionTypeWRF && p.ProposerElectionType != ProposerElectionTypeAbnormalRotation {
		return fmt.Errorf("unsupported proposer_election_type: %s", p.ProposerElectionType)
	}
	return nil
}

func (p *FinanceParams) Validate() error {
	if !p.MinGasPrice.ToBigInt().IsUint64() {
		return errors.Errorf("invalid min_gas_price %s, must be uint64", p.MinGasPrice)
	}
	return nil
}

func (p *StakeParams) Validate() error {
	if p.MaxAddStakeRatio == 0 {
		return errors.New("max_add_stake_ratio cannot be 0")
	}
	if p.MaxAddStakeRatio > RatioLimit {
		return errors.Errorf("max_add_stake_ratio cannot be greater than %d: %d", RatioLimit, p.MaxAddStakeRatio)
	}
	if p.MaxUnlockStakeRatio == 0 {
		return errors.New("max_unlock_stake_ratio cannot be 0")
	}
	if p.MaxUnlockStakeRatio > RatioLimit {
		return errors.Errorf("max_unlock_stake_ratio cannot be greater than %d: %d", RatioLimit, p.MaxUnlockStakeRatio)
	}
	if p.MaxUnlockingRecordNum == 0 {
		return errors.New("max_unlocking_record_num cannot be 0")
	}
	if p.MaxPendingInactiveValidatorRatio == 0 {
		return errors.New("max_pending_inactive_validator_ratio cannot be 0")
	}
	if p.MaxPendingInactiveValidatorRatio > RatioLimit {
		return errors.Errorf("max_pending_inactive_validator_ratio cannot be greater than %d: %d", RatioLimit, p.MaxPendingInactiveValidatorRatio)
	}

	if p.MinDelegateStake.ToBigInt().Cmp(new(big.Int)) < 0 {
		return errors.Errorf("invalid min_delegate_stake %s", p.MinDelegateStake)
	}
	if p.MinValidatorStake.ToBigInt().Cmp(new(big.Int)) < 0 {
		return errors.Errorf("invalid min_validator_stake %s", p.MinValidatorStake)
	}
	if p.MaxValidatorStake.ToBigInt().Cmp(new(big.Int)) <= 0 {
		return errors.Errorf("invalid max_validator_stake %s", p.MaxValidatorStake)
	}
	return nil
}

func (p *MiscParams) Validate() error {
	if p.TxMaxSize == 0 {
		return errors.New("tx_max_size cannot be 0")
	}
	return nil
}
