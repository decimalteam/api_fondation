package api_fondation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	web3 "github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/tendermint/tendermint/libs/log"

	rpc "github.com/tendermint/tendermint/rpc/client/http"
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

func NewWorker(cdc params.EncodingConfig, logger log.Logger, config *Config) (*Worker, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{}
	rpcClient, err := rpc.NewWithClient(config.RpcEndpoint, config.RpcEndpoint, httpClient)
	if err != nil {
		return nil, err
	}
	web3Client, err := web3.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}
	web3ChainId, err := web3Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	ethRpcClient, err := ethrpc.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}
	worker := &Worker{
		ctx:          context.Background(),
		httpClient:   httpClient,
		cdc:          cdc,
		logger:       logger,
		config:       config,
		hostname:     hostname,
		rpcClient:    rpcClient,
		web3Client:   web3Client,
		web3ChainId:  web3ChainId,
		ethRpcClient: ethRpcClient,
		query:        make(chan *ParseTask, 1000),
	}
	return worker, nil
}

func (w *Worker) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(w.config.WorkersCount)
	for i := 0; i < w.config.WorkersCount; i++ {
		go w.executeFromQuery(wg)
	}
	wg.Wait()
}

func (w *Worker) executeFromQuery(wg *sync.WaitGroup) {
	defer wg.Done()
	for {

		// Determine number of work to retrieve from the node
		w.getWork()

		// Retrieve block result from the node and prepare it for the indexer service
		task := <-w.query
		block := w.GetBlockResult(task.height, task.txNum)
		if block == nil {
			continue
		}

		// Send retrieved block result from the  indexer service
		data, err := json.Marshal(*block)
		w.panicError(err)
		w.sendBlock(task.height, data)
	}
}

func (w *Worker) GetBlockResult(height int64, txNum int) *Block {
	accum := NewEventAccumulator()

	w.logger.Info("Retrieving block results...", "block", height)

	// Fetch requested block from Tendermint RPC
	block := w.fetchBlock(height)
	if block == nil {
		return nil
	}

	// Fetch everything needed from Tendermint RPC aand EVM
	start := time.Now()
	txsChan := make(chan []Tx)
	resultsChan := make(chan *ctypes.ResultBlockResults)
	sizeChan := make(chan int)
	web3BlockChan := make(chan *web3types.Block)
	web3ReceiptsChan := make(chan web3types.Receipts)
	go w.fetchBlockResults(height, *block, accum, txsChan, resultsChan)
	go w.fetchBlockSize(height, sizeChan)
	txs := <-txsChan
	results := <-resultsChan
	size := <-sizeChan
	go w.fetchBlockWeb3(height, web3BlockChan)

	web3Block := <-web3BlockChan
	go w.fetchBlockTxReceiptsWeb3(web3Block, web3ReceiptsChan)
	web3Body := web3Block.Body()
	web3Transactions := make([]*TransactionEVM, len(web3Body.Transactions))
	for i, tx := range web3Body.Transactions {
		msg, err := tx.AsMessage(web3types.NewLondonSigner(w.web3ChainId), nil)
		w.panicError(err)
		web3Transactions[i] = &TransactionEVM{
			Type:             web3hexutil.Uint64(tx.Type()),
			Hash:             tx.Hash(),
			Nonce:            web3hexutil.Uint64(tx.Nonce()),
			BlockHash:        web3Block.Hash(),
			BlockNumber:      web3hexutil.Uint64(web3Block.NumberU64()),
			TransactionIndex: web3hexutil.Uint64(uint64(i)),
			From:             msg.From(),
			To:               msg.To(),
			Value:            (*web3hexutil.Big)(msg.Value()),
			Data:             web3hexutil.Bytes(msg.Data()),
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
			w.panicError(err)
		}
	}
	for _, event := range results.EndBlockEvents {
		err := accum.AddEvent(event, "")
		if err != nil {
			w.panicError(err)
		}
	}

	// TODO: move to event accumulator
	// Retrieve emission and rewards
	var emission string
	var rewards []ProposerReward
	var commissionRewards []CommissionReward
	for _, event := range results.EndBlockEvents {
		switch event.Type {
		case "emission":
			// Parse emission
			emission = string(event.Attributes[0].Value)

		case "proposer_reward":
			// Parse proposer rewards
			var reward ProposerReward
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
			var reward CommissionReward
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

	w.logger.Info(
		fmt.Sprintf("Compiled block (%s)", helpers.DurationToString(time.Since(start))),
		"block", height,
		"txs", len(txs),
		"begin-block-events", len(results.BeginBlockEvents),
		"end-block-events", len(results.EndBlockEvents),
	)

	// Create and fill Block object and then marshal to JSON
	return &Block{
		ID:                block.BlockID,
		Evidence:          block.Block.Evidence,
		Header:            block.Block.Header,
		LastCommit:        block.Block.LastCommit,
		Data:              BlockData{Txs: txs},
		Emission:          emission,
		Rewards:           rewards,
		CommissionRewards: commissionRewards,
		EndBlockEvents:    w.parseEvents(results.EndBlockEvents),
		BeginBlockEvents:  w.parseEvents(results.BeginBlockEvents),
		Size:              size,
		StateChanges:      *accum,
		EVM: BlockEVM{
			Header:       web3Block.Header(),
			Transactions: web3Transactions,
			Uncles:       web3Body.Uncles,
			Receipts:     web3Receipts,
		},
	}
}

func (w *Worker) panicError(err error) {
	if err != nil {
		w.logger.Error(fmt.Sprintf("Error: %v", err))
		panic(err)
	}
}

func (w *Worker) getWork() {
	// Uncomment for test purposes
	// w.query <- &ParseTask{
	// 	height: 319229,
	// 	txNum:  -1,
	// }
	// return

	// Prepare request
	url := fmt.Sprintf("%s/getWork", w.config.IndexerEndpoint)
	req, err := http.NewRequest("GET", url, nil)
	w.panicError(err)
	req.Header.Set("X-Worker", w.hostname)

	// Perform request
	resp, err := w.httpClient.Do(req)
	w.panicError(err)
	defer resp.Body.Close()

	// Parse response
	bodyBytes, err := io.ReadAll(resp.Body)
	w.panicError(err)
	height, err := strconv.Atoi(string(bodyBytes))
	w.panicError(err)

	// Send work to the channel
	w.query <- &ParseTask{
		height: int64(height),
		txNum:  -1,
	}
}

func (w *Worker) sendBlock(height int64, json []byte) {
	start := time.Now()

	// Prepare request
	url := fmt.Sprintf("%s/block", w.config.IndexerEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	w.panicError(err)
	req.Header.Set("X-Worker", w.hostname)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := w.httpClient.Do(req)
	w.panicError(err)

	// Parse response
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			w.logger.Error(
				fmt.Sprintf("Error: unable to send block to the indexer with status code: %s", resp.Status),
				"block", height,
			)
			w.resendBlock(height, json)
			return
		} else {
			w.logger.Info(
				fmt.Sprintf("Block is successfully sent (%s)", helpers.DurationToString(time.Since(start))),
				"block", height,
			)
			// Parse response
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				w.logger.Error(
					fmt.Sprintf("Error: unable to send block to the indexer: %s", err.Error()),
					"block", height,
				)
				w.logger.Info(
					fmt.Sprintf("Response from the indexer: %s", string(bodyBytes)),
					"block", height,
				)
				w.resendBlock(height, json)
				return
			}
		}
	}
	if err != nil {
		w.logger.Error(
			fmt.Sprintf("Error: unable to send block to the indexer: %s", err.Error()),
			"block", height,
		)
		w.resendBlock(height, json)
		return
	}
}

func (w *Worker) resendBlock(height int64, json []byte) {
	time.Sleep(time.Second)
	w.logger.Info("Retrying...", "block", height)
	w.sendBlock(height, json)
}

func (w *Worker) fetchBlock(height int64) *ctypes.ResultBlock {
	// Request until get block
	for first, start, deadline := true, time.Now(), time.Now().Add(RequestTimeout); true; first = false {
		// Request block
		result, err := w.rpcClient.Block(w.ctx, &height)
		if err == nil {
			if !first {
				w.logger.Info(
					fmt.Sprintf("Fetched block (after %s)", helpers.DurationToString(time.Since(start))),
					"block", height,
				)
			} else {
				w.logger.Info(
					fmt.Sprintf("Fetched block (%s)", helpers.DurationToString(time.Since(start))),
					"block", height,
				)
			}
			return result
		}
		// Stop trying when the deadline is reached
		if time.Now().After(deadline) {
			w.logger.Error("Failed to fetch block", "block", height, "error", err)
			return nil
		}
		// Sleep some time before next try
		time.Sleep(RequestRetryDelay)
	}

	return nil
}
