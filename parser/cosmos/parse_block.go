package cosmos

import (
	"bitbucket.org/decimalteam/api_fondation/clients"
	"context"
	"fmt"
	"math/big"
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

func Parse(ctx context.Context, blockNumber *big.Int) (*Block, error) {
	var res *Block

	web3Client, err := clients.GetWeb3Client(clients.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("get web3 client error: %v", err)
	}

	b, err := web3Client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("block by number error: %v", err)
	}

	dataTx, err := strconv.Atoi(string(b.Transaction(b.Hash()).Data()))
	if err != nil {
		return nil, fmt.Errorf("strconv atoi error: %v", err)
	}

	res = &Block{
		ID: b.Number().Int64(),
		Header: Header{
			Time:   strconv.Itoa(int(b.Header().Time)),
			Height: int(b.Number().Int64()),
		},
		Data: Data{
			Time:   strconv.FormatUint(b.Time(), 10),
			Height: int(b.Number().Int64()),
			DataTx: dataTx,
		},
	}

	return res, nil
}
