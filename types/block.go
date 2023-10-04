package types

import (
	web3types "github.com/ethereum/go-ethereum/core/types"
)

type ParseTask struct {
	height int64
	txNum  int
}

type BlockEVM struct {
	Header       *web3types.Header    `json:"header"`
	Transactions []*TransactionEVM    `json:"transactions"`
	Uncles       []*web3types.Header  `json:"uncles"`
	Receipts     []*web3types.Receipt `json:"receipts"`
}

//type Block struct {
//	ID                interface{}             `json:"id"`
//	Header            interface{}             `json:"header"`
//	Data              BlockData               `json:"data"`
//	Evidence          interface{}             `json:"evidence"`
//	LastCommit        interface{}             `json:"last_commit"`
//	Emission          string                  `json:"emission"`
//	Rewards           []ProposerReward        `json:"rewards"`
//	CommissionRewards []CommissionReward      `json:"commission_rewards"`
//	BeginBlockEvents  []Event                 `json:"begin_block_events"`
//	EndBlockEvents    []Event                 `json:"end_block_events"`
//	Size              int                     `json:"size"`
//	EVM               BlockEVM                `json:"evm"`
//	StateChanges      events.EventAccumulator `json:"state_changes"`
//}
//
//type BlockData struct {
//	Txs []Tx `json:"txs"`
//}
