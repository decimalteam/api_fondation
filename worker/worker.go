package worker

import (
	"bitbucket.org/decimalteam/api_fondation/clients"
	"context"
	"encoding/json"
	"fmt"
	"github.com/evmos/ethermint/encoding"
	"math/big"
	"net/http"
	"time"

	"bitbucket.org/decimalteam/api_fondation/events"
	"bitbucket.org/decimalteam/api_fondation/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	web3 "github.com/ethereum/go-ethereum/ethclient"
	"github.com/tendermint/tendermint/libs/log"
	rpc "github.com/tendermint/tendermint/rpc/client/http"

	//evmtypes "github.com/evmos/ethermint/x/evm/types"

	// Decimal modules

	//cointypes "bitbucket.org/decimalteam/go-smart-node/x/coin/types"
	//feetypes "bitbucket.org/decimalteam/go-smart-node/x/fee/types"
	//legacytypes "bitbucket.org/decimalteam/go-smart-node/x/legacy/types"
	//nfttypes "bitbucket.org/decimalteam/go-smart-node/x/nft/types"
	//swaptypes "bitbucket.org/decimalteam/go-smart-node/x/swap/types"
	//validatortypes "bitbucket.org/decimalteam/go-smart-node/x/validator/types"

	"github.com/dustin/go-humanize"
	"github.com/status-im/keycard-go/hexutils"
	abci "github.com/tendermint/tendermint/abci/types"

	web3common "github.com/ethereum/go-ethereum/common"
	web3hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	web3types "github.com/ethereum/go-ethereum/core/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

type ParseTask struct {
	height int64
	txNum  int
}

func GetBlockResult(height int64) *types.Block {
	ctx := context.Background()

	accum := events.NewEventAccumulator()

	fmt.Printf("Retrieving block results...", "block", height)

	cdc := encoding.MakeConfig(getModuleBasics())

	rpcClient, err := clients.GetRpcClient(getConfig(), clients.GetHttpClient())
	if err != nil {
		panicError(err)
	}

	web3Client, err := clients.GetWeb3Client(getConfig())
	if err != nil {
		panicError(err)
	}

	ethRpcClient, err := clients.GetEthRpcClient(getConfig())
	if err != nil {
		panicError(err)
	}

	// Fetch requested block from Tendermint RPC
	block := fetchBlock(ctx, rpcClient, height)
	if block == nil {
		return nil
	}

	// Fetch everything needed from Tendermint RPC aand EVM
	start := time.Now()
	txsChan := make(chan []types.Tx)
	resultsChan := make(chan *ctypes.ResultBlockResults)
	sizeChan := make(chan int)
	web3BlockChan := make(chan *web3types.Block)
	web3ReceiptsChan := make(chan web3types.Receipts)
	go fetchBlockResults(ctx, rpcClient, cdc, height, *block, accum, txsChan, resultsChan)
	go fetchBlockSize(ctx, rpcClient, height, sizeChan)
	txs := <-txsChan
	results := <-resultsChan
	size := <-sizeChan
	go fetchBlockWeb3(ctx, web3Client, height, web3BlockChan)

	web3Block := <-web3BlockChan
	go fetchBlockTxReceiptsWeb3(ethRpcClient, web3Block, web3ReceiptsChan)
	web3Body := web3Block.Body()
	web3Transactions := make([]*types.TransactionEVM, len(web3Body.Transactions))
	for i, tx := range web3Body.Transactions {
		web3Client, err := clients.GetWeb3Client(getConfig())
		if err != nil {
			panicError(err)
		}

		web3ChainId, err := clients.GetWeb3ChainId(web3Client)
		if err != nil {
			panicError(err)
		}

		msg, err := tx.AsMessage(web3types.NewLondonSigner(web3ChainId), nil)
		panicError(err)
		web3Transactions[i] = &types.TransactionEVM{
			Type:             web3hexutil.Uint64(tx.Type()),
			Hash:             tx.Hash(),
			Nonce:            web3hexutil.Uint64(tx.Nonce()),
			BlockHash:        web3Block.Hash(),
			BlockNumber:      web3hexutil.Uint64(web3Block.NumberU64()),
			TransactionIndex: web3hexutil.Uint64(uint64(i)),
			From:             msg.From(),
			To:               msg.To(),
			Value:            (*web3hexutil.Big)(msg.Value()),
			Data:             msg.Data(),
			Gas:              web3hexutil.Uint64(msg.Gas()),
			GasPrice:         (*web3hexutil.Big)(msg.GasPrice()),
			ChainId:          (*web3hexutil.Big)(tx.ChainId()),
			AccessList:       msg.AccessList(),
			GasTipCap:        (*web3hexutil.Big)(msg.GasTipCap()),
			GasFeeCap:        (*web3hexutil.Big)(msg.GasFeeCap()),
		}
	}
	web3Receipts := <-web3ReceiptsChan

	for _, event := range results.BeginBlockEvents {
		err := accum.AddEvent(event, "")
		if err != nil {
			panicError(err)
		}
	}
	for _, event := range results.EndBlockEvents {
		err := accum.AddEvent(event, "")
		if err != nil {
			panicError(err)
		}
	}

	// TODO: move to event accumulator
	// Retrieve emission and rewards
	var emission string
	var rewards []types.ProposerReward
	var commissionRewards []types.CommissionReward
	for _, event := range results.EndBlockEvents {
		switch event.Type {
		case "emission":
			// Parse emission
			emission = string(event.Attributes[0].Value)

		case "proposer_reward":
			// Parse proposer rewards
			var reward types.ProposerReward
			for _, attr := range event.Attributes {
				switch string(attr.Key) {
				case "amount", "accum_rewards":
					reward.Reward = string(attr.Value)
				case "validator", "accum_rewards_validator":
					reward.Address = string(attr.Value)
				case "delegator":
					reward.Delegator = string(attr.Value)
				}
			}
			rewards = append(rewards, reward)

		case "commission_reward":
			// Parser commission reward
			var reward types.CommissionReward
			for _, attr := range event.Attributes {
				switch string(attr.Key) {
				case "amount":
					reward.Amount = string(attr.Value)
				case "validator":
					reward.Validator = string(attr.Value)
				case "reward_address":
					reward.RewardAddress = string(attr.Value)
				}
			}
			commissionRewards = append(commissionRewards, reward)
		}
	}

	fmt.Println(
		fmt.Sprintf("Compiled block (%s)", DurationToString(time.Since(start))),
		"block", height,
		"txs", len(txs),
		"begin-block-events", len(results.BeginBlockEvents),
		"end-block-events", len(results.EndBlockEvents),
	)

	// Create and fill Block object and then marshal to JSON
	return &types.Block{
		ID:                block.BlockID,
		Evidence:          block.Block.Evidence,
		Header:            block.Block.Header,
		LastCommit:        block.Block.LastCommit,
		Data:              types.BlockData{Txs: txs},
		Emission:          emission,
		Rewards:           rewards,
		CommissionRewards: commissionRewards,
		EndBlockEvents:    parseEvents(results.EndBlockEvents),
		BeginBlockEvents:  parseEvents(results.BeginBlockEvents),
		Size:              size,
		StateChanges:      *accum,
		EVM: types.BlockEVM{
			Header:       web3Block.Header(),
			Transactions: web3Transactions,
			Uncles:       web3Body.Uncles,
			Receipts:     web3Receipts,
		},
	}
}

func panicError(err error) {
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %v", err))
		panic(err)
	}
}

func fetchBlock(ctx context.Context, rpcClient *rpc.HTTP, height int64) *ctypes.ResultBlock {
	// Request until get block
	for first, start, deadline := true, time.Now(), time.Now().Add(types.RequestTimeout); true; first = false {
		// Request block
		result, err := rpcClient.Block(ctx, &height)
		if err == nil {
			if !first {
				fmt.Println(
					fmt.Sprintf("Fetched block (after %s)", DurationToString(time.Since(start))),
					"block", height,
				)
			} else {
				fmt.Println(
					fmt.Sprintf("Fetched block (%s)", DurationToString(time.Since(start))),
					"block", height,
				)
			}
			return result
		}
		// Stop trying when the deadline is reached
		if time.Now().After(deadline) {
			fmt.Printf("Failed to fetch block", "block", height, "error\n", err)
			return nil
		}
		// Sleep some time before next try
		time.Sleep(types.RequestRetryDelay)
	}

	return nil
}

func fetchBlockSize(ctx context.Context, rpcClient *rpc.HTTP, height int64, ch chan int) {

	// Request blockchain info
	result, err := rpcClient.BlockchainInfo(ctx, height, height)
	panicError(err)

	// Send result to the channel
	ch <- result.BlockMetas[0].BlockSize
}

func fetchBlockResults(ctx context.Context, rpcClient *rpc.HTTP, cdc params.EncodingConfig, height int64, block ctypes.ResultBlock, ea *events.EventAccumulator, ch chan []types.Tx, brch chan *ctypes.ResultBlockResults) {
	var err error

	// Request block results from the node
	// NOTE: Try to retrieve results in the loop since it looks like there is some delay before results are ready to by retrieved
	var blockResults *ctypes.ResultBlockResults
	for c := 1; true; c++ {
		if c > 5 {
			fmt.Println(fmt.Sprintf("%d attempt to fetch block height: %d, time %s", c, height, time.Now().String()))
		}
		// Request block results
		blockResults, err = rpcClient.BlockResults(ctx, &height)
		if err == nil {
			break
		}
		// Sleep some time before next try
		time.Sleep(types.RequestRetryDelay)
	}

	// Prepare block results by overall processing
	var results []types.Tx
	for i, tx := range block.Block.Txs {
		var result types.Tx
		var txLog []interface{}
		txr := blockResults.TxsResults[i]

		recoveredTx, err := cdc.TxConfig.TxDecoder()(tx)
		panicError(err)

		// Parse transaction results logs
		err = json.Unmarshal([]byte(txr.Log), &txLog)
		if err != nil {
			result.Log = []interface{}{types.FailedTxLog{Log: txr.Log}}
		} else {
			result.Log = txLog
		}

		result.Info = parseTxInfo(recoveredTx, cdc)
		result.Data = txr.Data
		result.Hash = hexutils.BytesToHex(tx.Hash())
		result.Code = txr.Code
		result.Codespace = txr.Codespace
		result.GasUsed = txr.GasUsed
		result.GasWanted = txr.GasWanted

		results = append(results, result)

		// process events for transactions
		for _, event := range txr.Events {
			err := ea.AddEvent(event, hexutils.BytesToHex(tx.Hash()))
			if err != nil {
				fmt.Printf("error in event %v\n", event.Type)
				panicError(err)
			}
		}
	}

	// Send results to the channel
	ch <- results
	brch <- blockResults
}

func fetchBlockWeb3(ctx context.Context, web3Client *web3.Client, height int64, ch chan *web3types.Block) {

	// Request block by number
	result, err := web3Client.BlockByNumber(ctx, big.NewInt(height))
	panicError(err)

	// Send result to the channel
	ch <- result
}

func fetchBlockTxReceiptsWeb3(ethRpcClient *ethrpc.Client, block *web3types.Block, ch chan web3types.Receipts) {
	txCount := len(block.Transactions())
	results := make(web3types.Receipts, txCount)
	requests := make([]ethrpc.BatchElem, txCount)
	if txCount == 0 {
		ch <- results
		return
	}

	// NOTE: Try to retrieve tx receipts in the loop since it looks like there is some delay before receipts are ready to by retrieved
	for c := 1; true; c++ {
		if c > 5 {
			fmt.Println(fmt.Sprintf("%d attempt to fetch transaction receipts with height: %d, time %s", c, block.NumberU64(), time.Now().String()))
		}
		// Prepare batch requests to retrieve the receipt for each transaction in the block
		for i, tx := range block.Transactions() {
			results[i] = &web3types.Receipt{}
			requests[i] = ethrpc.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{tx.Hash()},
				Result: results[i],
			}
		}
		// Request transaction receipts with a batch
		err := ethRpcClient.BatchCall(requests[:])
		if err == nil {
			// Ensure all transaction receipts are retrieved
			for i := range requests {
				txHash := requests[i].Args[0].(web3common.Hash)
				if requests[i].Error != nil {
					err = requests[i].Error
					if c > 5 {
						fmt.Println(fmt.Sprintf("Error: %v", err))
					}
					continue
				}
				if results[i].BlockNumber == nil || results[i].BlockNumber.Sign() == 0 {
					err = fmt.Errorf("got null result for tx with hash %v", txHash)
					if c > 5 {
						fmt.Println(fmt.Sprintf("Error: %v", err))
					}
				}
			}
		}
		if err == nil {
			break
		}
		// Sleep some time before next try
		time.Sleep(types.RequestRetryDelay)
	}

	// Send results to the channel
	ch <- results
}

// DurationToString converts provided duration to human readable string presentation.
func DurationToString(d time.Duration) string {
	ns := time.Duration(d.Nanoseconds())
	ms := float64(ns) / 1000000.0
	var unit string
	var amount string
	if ns < time.Microsecond {
		amount, unit = humanize.CommafWithDigits(float64(ns), 0), "ns"
	} else if ns < time.Millisecond {
		amount, unit = humanize.CommafWithDigits(ms*1000.0, 3), "μs"
	} else if ns < time.Second {
		amount, unit = humanize.CommafWithDigits(ms, 3), "ms"
	} else if ns < time.Minute {
		amount, unit = humanize.CommafWithDigits(ms/1000.0, 3), "s"
	} else if ns < time.Hour {
		amount, unit = humanize.CommafWithDigits(ms/60000.0, 3), "m"
	} else if ns < 24*time.Hour {
		amount, unit = humanize.CommafWithDigits(ms/3600000.0, 3), "h"
	} else {
		days := ms / 86400000.0
		unit = "day"
		if days > 1 {
			unit = "days"
		}
		amount = humanize.CommafWithDigits(days, 3)
	}
	return fmt.Sprintf("%s %s", amount, unit)
}

func parseTxInfo(tx sdk.Tx, cdc params.EncodingConfig) (txInfo types.TxInfo) {
	if tx == nil {
		return
	}
	for _, rawMsg := range tx.GetMsgs() {
		params := make(map[string]interface{})
		err := json.Unmarshal(cdc.Codec.MustMarshalJSON(rawMsg), &params)
		panicError(err)
		var msg types.TxMsg
		msg.Type = sdk.MsgTypeURL(rawMsg)
		msg.Params = params
		for _, signer := range rawMsg.GetSigners() {
			msg.From = append(msg.From, signer.String())
		}
		txInfo.Msgs = append(txInfo.Msgs, msg)
	}
	txInfo.Fee.Gas = tx.(sdk.FeeTx).GetGas()
	txInfo.Fee.Amount = tx.(sdk.FeeTx).GetFee()
	txInfo.Memo = tx.(sdk.TxWithMemo).GetMemo()
	return
}

func parseEvents(events []abci.Event) []types.Event {
	var newEvents []types.Event
	for _, ev := range events {
		newEvent := types.Event{
			Type:       ev.Type,
			Attributes: []types.Attribute{},
		}
		for _, attr := range ev.Attributes {
			newEvent.Attributes = append(newEvent.Attributes, types.Attribute{
				Key:   string(attr.Key),
				Value: string(attr.Value),
			})
		}
		newEvents = append(newEvents, newEvent)
	}
	return newEvents
}
