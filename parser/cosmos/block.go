package cosmos

import (
	"bitbucket.org/decimalteam/api_fondation/worker"
)

type Block struct {
	ID                BlockId            `json:"id"`
	Header            Header             `json:"header"`
	Data              BlockTx            `json:"data"`
	Evidence          interface{}        `json:"evidence"`
	LastCommit        interface{}        `json:"last_commit"`
	Emission          string             `json:"emission"`
	Rewards           []ProposerReward   `json:"rewards"`
	CommissionRewards []CommissionReward `json:"commission_rewards"`
	BeginBlockEvents  []Event            `json:"begin_block_events"`
	EndBlockEvents    []Event            `json:"end_block_events"`
	Size              int                `json:"size"`
}

type BlockTx struct {
	Txs []Tx `json:"txs"`
}

type ProposerReward struct {
	Reward    string `json:"reward"`
	Address   string `json:"address"`
	Delegator string `json:"delegator"`
}

type CommissionReward struct {
	Amount        string `json:"amount"`
	Validator     string `json:"validator"`
	RewardAddress string `json:"reward_address"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Header struct {
	Time   string `json:"time"`
	Height int    `json:"height"`
}

type BlockId struct {
	Hash string `json:"hash"`
}

func Parse(blockNumber *int64) (*Block, error) {
	return worker.GetBlockResult(*blockNumber), nil
}
