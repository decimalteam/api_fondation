package cosmos

import (
	"bitbucket.org/decimalteam/api_fondation/clients"
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Block struct {
	ID                interface{}        `json:"id"`
	Header            interface{}        `json:"header"`
	Data              BlockData          `json:"data"`
	Evidence          interface{}        `json:"evidence"`
	LastCommit        interface{}        `json:"last_commit"`
	Emission          string             `json:"emission"`
	Rewards           []ProposerReward   `json:"rewards"`
	CommissionRewards []CommissionReward `json:"commission_rewards"`
	BeginBlockEvents  []Event            `json:"begin_block_events"`
	EndBlockEvents    []Event            `json:"end_block_events"`
	Size              int                `json:"size"`
}

type BlockData struct {
	Txs []Tx `json:"txs"`
}

type Tx struct {
	Hash      string        `json:"hash"`
	Log       []interface{} `json:"log"`
	Code      uint32        `json:"code"`
	Codespace string        `json:"codespace"`
	Data      interface{}   `json:"data"`
	GasUsed   int64         `json:"gas_used"`
	GasWanted int64         `json:"gas_wanted"`
	Info      TxInfo        `json:"info"`
}

type FailedTxLog struct {
	Log string `json:"log"`
}

type TxMsg struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
	From   []string    `json:"from"`
}

type TxFee struct {
	Gas    uint64    `json:"gas"`
	Amount sdk.Coins `json:"amount"`
}

type TxInfo struct {
	Msgs []TxMsg `json:"msgs"`
	Memo string  `json:"memo"`
	Fee  TxFee   `json:"fee"`
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
