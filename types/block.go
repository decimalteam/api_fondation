package types

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	web3 "github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/tendermint/tendermint/libs/log"
	rpc "github.com/tendermint/tendermint/rpc/client/http"
	"math/big"
	"net/http"
)

type Worker struct {
	ctx          context.Context
	httpClient   *http.Client
	cdc          params.EncodingConfig
	logger       log.Logger
	config       *Config
	hostname     string
	rpcClient    *rpc.HTTP
	web3Client   *web3.Client
	web3ChainId  *big.Int
	ethRpcClient *ethrpc.Client
	query        chan *ParseTask
}

type Config struct {
	IndexerEndpoint string
	RpcEndpoint     string
	Web3Endpoint    string
	WorkersCount    int
}

type ParseTask struct {
	height int64
	txNum  int
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
//
//type BlockEVM struct {
//	Header       *web3types.Header    `json:"header"`
//	Transactions []*TransactionEVM    `json:"transactions"`
//	Uncles       []*web3types.Header  `json:"uncles"`
//	Receipts     []*web3types.Receipt `json:"receipts"`
//}
