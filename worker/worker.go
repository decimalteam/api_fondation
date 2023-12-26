package worker

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/events"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	web3 "github.com/ethereum/go-ethereum/ethclient"
	"github.com/evmos/ethermint/encoding"
	rpc "github.com/tendermint/tendermint/rpc/client/http"

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

//func GetAllTransactionByHeights(height int64, withTrx bool) *types.BlockData {
//	ctx := context.Background()
//
//	accum := events.NewEventAccumulator()
//
//	cdc := encoding.MakeConfig(GetModuleBasics())
//
//	cl, err := client.New()
//	if err != nil {
//		panicError(err)
//	}
//
//	// Fetch requested block from Tendermint RPC
//	block := fetchBlock(ctx, cl.RpcClient, height)
//	if block == nil {
//		return nil
//	}
//
//	// Fetch everything needed from Tendermint RPC and EVM
//	start := time.Now()
//	txsChan := make(chan []cosmos.Tx)
//	resultsChan := make(chan *ctypes.ResultBlockResults)
//	sizeChan := make(chan int)
//	web3BlockChan := make(chan *web3types.Block)
//	web3ReceiptsChan := make(chan web3types.Receipts)
//	go fetchBlockResults(ctx, cl.RpcClient, cdc, height, *block, accum, txsChan, resultsChan)
//	go fetchBlockSize(ctx, cl.RpcClient, height, sizeChan)
//	txs := <-txsChan
//	results := <-resultsChan
//	size := <-sizeChan
//	go FetchBlockWeb3(ctx, cl.Web3Client, height, web3BlockChan)
//
//	web3Block := <-web3BlockChan
//	web3Body := web3Block.Body()
//	web3Transactions := make([]*evm.TransactionEVM, len(web3Body.Transactions))
//	if withTrx {
//		go FetchBlockTxReceiptsWeb3(cl.EthRpcClient, web3Block, web3ReceiptsChan)
//		for i, tx := range web3Body.Transactions {
//			msg, err := tx.AsMessage(web3types.NewLondonSigner(cl.Web3ChainId), nil)
//			panicError(err)
//			web3Transactions[i] = &evm.TransactionEVM{
//				Type:             web3hexutil.Uint64(tx.Type()),
//				Hash:             tx.Hash(),
//				Nonce:            web3hexutil.Uint64(tx.Nonce()),
//				BlockHash:        web3Block.Hash(),
//				BlockNumber:      web3hexutil.Uint64(web3Block.NumberU64()),
//				TransactionIndex: web3hexutil.Uint64(uint64(i)),
//				From:             msg.From(),
//				To:               msg.To(),
//				Value:            (*web3hexutil.Big)(msg.Value()),
//				Data:             msg.Data(),
//				Gas:              web3hexutil.Uint64(msg.Gas()),
//				GasPrice:         (*web3hexutil.Big)(msg.GasPrice()),
//				ChainId:          (*web3hexutil.Big)(tx.ChainId()),
//				AccessList:       msg.AccessList(),
//				GasTipCap:        (*web3hexutil.Big)(msg.GasTipCap()),
//				GasFeeCap:        (*web3hexutil.Big)(msg.GasFeeCap()),
//			}
//		}
//	}
//
//	fmt.Println(
//		fmt.Sprintf("Compiled block (%s)", DurationToString(time.Since(start))),
//		"block", height,
//		"txs", len(txs),
//		"begin-block-events", len(results.BeginBlockEvents),
//		"end-block-events", len(results.EndBlockEvents),
//	)
//
//	web3Receipts := <-web3ReceiptsChan
//
//	// Create and fill Block object and then marshal to JSON
//	return &types.BlockData{
//		CosmosBlock: &cosmos.Block{
//			ID:       cosmos.BlockId{Hash: block.BlockID.Hash.String()},
//			Evidence: block.Block.Evidence,
//			Header: cosmos.Header{
//				Time:   block.Block.Time.String(),
//				Height: int(block.Block.Height),
//			},
//			LastCommit:       block.Block.LastCommit,
//			Data:             cosmos.BlockTx{Txs: txs},
//			EndBlockEvents:   parseEvents(results.EndBlockEvents),
//			BeginBlockEvents: parseEvents(results.BeginBlockEvents),
//			Size:             size,
//		},
//		EvmBlock: &evm.BlockEVM{
//			Header:       web3Block.Header(),
//			Transactions: web3Transactions,
//			Uncles:       web3Body.Uncles,
//			Receipts:     web3Receipts,
//		},
//	}
//}

func GetEvmBlock(height int64) *types.BlockData {
	ctx := context.Background()

	cl, err := client.New()
	if err != nil {
		panicError(err)
		return nil
	}

	// Fetch everything needed from Tendermint RPC and EVM
	start := time.Now()
	web3BlockChan := make(chan *web3types.Block)
	web3ReceiptsChan := make(chan web3types.Receipts)
	go FetchBlockWeb3(ctx, cl.Web3Client, height, web3BlockChan)

	web3Block := <-web3BlockChan
	web3Body := web3Block.Body()
	web3Transactions := make([]*evm.TransactionEVM, len(web3Body.Transactions))

	go FetchBlockTxReceiptsWeb3(cl.EthRpcClient, web3Block, web3ReceiptsChan)
	for i, transaction := range web3Body.Transactions {
		msg, err := transaction.AsMessage(web3types.NewLondonSigner(cl.Web3ChainId), nil)
		if err != nil {
			panicError(err)
			continue
		}
		web3Transactions[i] = &evm.TransactionEVM{
			Type:             web3hexutil.Uint64(transaction.Type()),
			Hash:             transaction.Hash(),
			Nonce:            web3hexutil.Uint64(transaction.Nonce()),
			BlockHash:        web3Block.Hash(),
			BlockNumber:      web3hexutil.Uint64(web3Block.NumberU64()),
			TransactionIndex: web3hexutil.Uint64(uint64(i)),
			From:             msg.From(),
			To:               msg.To(),
			Value:            (*web3hexutil.Big)(msg.Value()),
			Data:             msg.Data(),
			Gas:              web3hexutil.Uint64(msg.Gas()),
			GasPrice:         (*web3hexutil.Big)(msg.GasPrice()),
			ChainId:          (*web3hexutil.Big)(transaction.ChainId()),
			AccessList:       msg.AccessList(),
			GasTipCap:        (*web3hexutil.Big)(msg.GasTipCap()),
			GasFeeCap:        (*web3hexutil.Big)(msg.GasFeeCap()),
		}
	}

	fmt.Println(
		fmt.Sprintf("Compiled block (%s)", DurationToString(time.Since(start))),
		"block", height,
		"txs", len(web3Transactions),
	)

	web3Receipts := <-web3ReceiptsChan

	// Create and fill Block object and then marshal to JSON
	return &types.BlockData{
		CosmosBlock: &cosmos.Block{},
		EvmBlock: &evm.BlockEVM{
			Header:       web3Block.Header(),
			Transactions: web3Transactions,
			Uncles:       web3Body.Uncles,
			Receipts:     web3Receipts,
		},
	}
}

func GetBlockOnly(height int64) *types.BlockData {
	ctx := context.Background()

	accum := events.NewEventAccumulator()

	cdc := encoding.MakeConfig(GetModuleBasics())

	cl, err := client.New()
	if err != nil {
		panicError(err)
		return nil
	}

	// Fetch requested block from Tendermint RPC
	block := fetchBlock(ctx, cl.RpcClient, height)
	if block == nil {
		return nil
	}

	// Fetch everything needed from Tendermint RPC and EVM
	start := time.Now()
	txsChan := make(chan []cosmos.Tx)
	resultsChan := make(chan *ctypes.ResultBlockResults)
	sizeChan := make(chan int)
	web3BlockChan := make(chan *web3types.Block)
	go fetchBlockResults(ctx, cl.RpcClient, cdc, height, *block, accum, txsChan, resultsChan)
	go fetchBlockSize(ctx, cl.RpcClient, height, sizeChan)
	txs := <-txsChan
	results := <-resultsChan
	size := <-sizeChan
	go FetchBlockWeb3(ctx, cl.Web3Client, height, web3BlockChan)

	web3Block := <-web3BlockChan
	if web3Block == nil {
		return nil
	}
	web3Body := web3Block.Body()
	web3Transactions := make([]*evm.TransactionEVM, len(web3Body.Transactions))

	fmt.Println(
		fmt.Sprintf("Compiled block (%s)", DurationToString(time.Since(start))),
		"block", height,
		"txs", len(txs),
		"begin-block-events", len(results.BeginBlockEvents),
		"end-block-events", len(results.EndBlockEvents),
	)

	// Create and fill Block object and then marshal to JSON
	return &types.BlockData{
		CosmosBlock: &cosmos.Block{
			ID:       cosmos.BlockId{Hash: block.BlockID.Hash.String()},
			Evidence: block.Block.Evidence,
			Header: cosmos.Header{
				Time:   block.Block.Time.String(),
				Height: int(block.Block.Height),
			},
			LastCommit: block.Block.LastCommit,
			Data:       cosmos.BlockTx{Txs: txs},
			Size:       size,
		},
		EvmBlock: &evm.BlockEVM{
			Header:       web3Block.Header(),
			Transactions: web3Transactions,
			Uncles:       web3Body.Uncles,
		},
	}
}

func GetBlockResult(height int64) *types.BlockData {
	ctx := context.Background()

	accum := events.NewEventAccumulator()

	cdc := encoding.MakeConfig(GetModuleBasics())

	cl, err := client.New()
	if err != nil {
		panicError(err)
		return nil
	}

	// Fetch requested block from Tendermint RPC
	block := fetchBlock(ctx, cl.RpcClient, height)
	if block == nil {
		return nil
	}

	// Fetch everything needed from Tendermint RPC and EVM
	start := time.Now()
	txsChan := make(chan []cosmos.Tx)
	resultsChan := make(chan *ctypes.ResultBlockResults)
	sizeChan := make(chan int)
	web3BlockChan := make(chan *web3types.Block)
	web3ReceiptsChan := make(chan web3types.Receipts)
	go fetchBlockResults(ctx, cl.RpcClient, cdc, height, *block, accum, txsChan, resultsChan)
	go fetchBlockSize(ctx, cl.RpcClient, height, sizeChan)
	txs := <-txsChan
	results := <-resultsChan
	size := <-sizeChan
	go FetchBlockWeb3(ctx, cl.Web3Client, height, web3BlockChan)

	web3Block := <-web3BlockChan
	web3Body := web3Block.Body()
	web3Transactions := make([]*evm.TransactionEVM, len(web3Body.Transactions))

	go FetchBlockTxReceiptsWeb3(cl.EthRpcClient, web3Block, web3ReceiptsChan)
	for i, tx := range web3Body.Transactions {
		msg, err := tx.AsMessage(web3types.NewLondonSigner(cl.Web3ChainId), nil)
		if err != nil {
			panicError(err)
			continue
		}
		web3Transactions[i] = &evm.TransactionEVM{
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

	for _, event := range results.BeginBlockEvents {
		err := accum.AddEvent(event, "")
		if err != nil {
			panicError(err)
			continue
		}
	}
	for _, event := range results.EndBlockEvents {
		err := accum.AddEvent(event, "")
		if err != nil {
			panicError(err)
			continue
		}
	}

	// TODO: move to event accumulator
	// Retrieve emission and rewards
	var emission string
	var rewards []cosmos.ProposerReward
	var commissionRewards []cosmos.CommissionReward
	for _, event := range results.EndBlockEvents {
		switch event.Type {
		case "emission":
			// Parse emission
			emission = string(event.Attributes[0].Value)

		case "proposer_reward":
			// Parse proposer rewards
			var reward cosmos.ProposerReward
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
			var reward cosmos.CommissionReward
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

	web3Receipts := <-web3ReceiptsChan

	// Create and fill Block object and then marshal to JSON
	return &types.BlockData{
		CosmosBlock: &cosmos.Block{
			ID:       cosmos.BlockId{Hash: block.BlockID.Hash.String()},
			Evidence: block.Block.Evidence,
			Header: cosmos.Header{
				Time:            block.Block.Time.String(),
				Height:          int(block.Block.Height),
				ProposerAddress: block.Block.ProposerAddress,
			},
			LastCommit:        block.Block.LastCommit,
			Data:              cosmos.BlockTx{Txs: txs},
			Emission:          emission,
			Rewards:           rewards,
			CommissionRewards: commissionRewards,
			EndBlockEvents:    parseEvents(results.EndBlockEvents),
			BeginBlockEvents:  parseEvents(results.BeginBlockEvents),
			Size:              size,
			StateChanges:      *accum,
		},
		EvmBlock: &evm.BlockEVM{
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
		return
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
	if err != nil {
		panicError(err)
		return
	}

	// Send result to the channel
	ch <- result.BlockMetas[0].BlockSize
}

func fetchBlockResults(ctx context.Context, rpcClient *rpc.HTTP, cdc params.EncodingConfig, height int64, block ctypes.ResultBlock, ea *events.EventAccumulator, ch chan []cosmos.Tx, brch chan *ctypes.ResultBlockResults) {
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
	var results []cosmos.Tx
	for i, tx := range block.Block.Txs {
		var result cosmos.Tx
		var txLog []interface{}
		txr := blockResults.TxsResults[i]

		recoveredTx, err := cdc.TxConfig.TxDecoder()(tx)
		if err != nil {
			panicError(err)
			continue
		}

		// Parse transaction results logs
		err = json.Unmarshal([]byte(txr.Log), &txLog)
		if err != nil {
			result.Log = []interface{}{cosmos.FailedTxLog{Log: txr.Log}}
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
				continue
			}
		}
	}

	// Send results to the channel
	ch <- results
	brch <- blockResults
}

func FetchBlockWeb3(ctx context.Context, web3Client *web3.Client, height int64, ch chan *web3types.Block) {

	// Request block by number
	result, err := web3Client.BlockByNumber(ctx, big.NewInt(height))
	if err != nil {
		panicError(err)
	}

	// Send result to the channel
	ch <- result
}

func FetchBlockTxReceiptsWeb3(ethRpcClient *ethrpc.Client, block *web3types.Block, ch chan web3types.Receipts) {
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

func parseTxInfo(tx sdk.Tx, cdc params.EncodingConfig) (txInfo cosmos.TxInfo) {
	if tx == nil {
		return
	}
	for _, rawMsg := range tx.GetMsgs() {
		parameters := make(map[string]interface{})
		err := json.Unmarshal(cdc.Codec.MustMarshalJSON(rawMsg), &parameters)
		if err != nil {
			panicError(err)
			continue
		}
		var msg cosmos.TxMsg
		msg.Type = sdk.MsgTypeURL(rawMsg)
		msg.Params = parameters
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

func parseEvents(events []abci.Event) []cosmos.Event {
	var newEvents []cosmos.Event
	for _, ev := range events {
		newEvent := cosmos.Event{
			Type:       ev.Type,
			Attributes: []cosmos.Attribute{},
		}
		for _, attr := range ev.Attributes {
			newEvent.Attributes = append(newEvent.Attributes, cosmos.Attribute{
				Key:   string(attr.Key),
				Value: string(attr.Value),
			})
		}
		newEvents = append(newEvents, newEvent)
	}
	return newEvents
}
