package cosmos

import (
	"bitbucket.org/decimalteam/api_fondation/types"
	"strconv"
)

type Block struct {
	ID     interface{} `json:"id"`
	Header Header      `json:"header"`
	Data   Data        `json:"data"`
}

type Data struct {
	Time              string `json:"time"`
	Height            int    `json:"height"`
	DataTx            int    `json:"data"`
	Header            Header `json:"header"`
	Rewards           int    `json:"rewards"`
	Emission          int    `json:"emission"`
	Evidence          int    `json:"evidence"`
	LastCommit        int    `json:"last_commit"`
	StageChange       int    `json:"stage_change"`
	EndBlockEvents    int    `json:"end_block_events"`
	BeginBlockEvents  int    `json:"begin_block_events"`
	CommissionRewards int    `json:"commission_rewards"`
}

type Header struct {
	Time   string `json:"time"`
	Height int    `json:"height"`
}

func Parse(block Block) *types.Block {
	var res *types.Block

	res = &types.Block{
		ID: block.ID,
		Header: Header{
			Time:   block.Header.Time,
			Height: block.Header.Height,
		},
		Evidence:   block.Data.Evidence,
		LastCommit: block.Data.LastCommit,
		Emission:   strconv.Itoa(block.Data.Emission),
		Rewards: []types.ProposerReward{
			{Reward: strconv.Itoa(block.Data.Rewards)},
		},
		CommissionRewards: []types.CommissionReward{
			{Amount: strconv.Itoa(block.Data.CommissionRewards)},
		},
		BeginBlockEvents: []types.Event{
			{Type: strconv.Itoa(block.Data.BeginBlockEvents)},
		},
		EndBlockEvents: []types.Event{
			{Type: strconv.Itoa(block.Data.EndBlockEvents)},
		},
	}

	return res
}
