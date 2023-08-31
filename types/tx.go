package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	web3common "github.com/ethereum/go-ethereum/common"
	web3hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	web3types "github.com/ethereum/go-ethereum/core/types"
)

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

type FailedTxLog struct {
	Log string `json:"log"`
}

type TransactionEVM struct {
	Type             web3hexutil.Uint64  `json:"type"`
	Hash             web3common.Hash     `json:"hash"`
	Nonce            web3hexutil.Uint64  `json:"nonce"`
	BlockHash        web3common.Hash     `json:"blockHash"`
	BlockNumber      web3hexutil.Uint64  `json:"blockNumber"`
	TransactionIndex web3hexutil.Uint64  `json:"transactionIndex"`
	From             web3common.Address  `json:"from"`
	To               *web3common.Address `json:"to"`
	Value            *web3hexutil.Big    `json:"value"`
	Data             web3hexutil.Bytes   `json:"input"`
	Gas              web3hexutil.Uint64  `json:"gas"`
	GasPrice         *web3hexutil.Big    `json:"gasPrice"`

	// Optional
	ChainId    *web3hexutil.Big     `json:"chainId,omitempty"`              // EIP-155 replay protection
	AccessList web3types.AccessList `json:"accessList,omitempty"`           // EIP-2930 access list
	GasTipCap  *web3hexutil.Big     `json:"maxPriorityFeePerGas,omitempty"` // EIP-1559 dynamic fee transactions
	GasFeeCap  *web3hexutil.Big     `json:"maxFeePerGas,omitempty"`         // EIP-1559 dynamic fee transactions
}
