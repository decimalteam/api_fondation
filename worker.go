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

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	// Decimal modules

	cointypes "bitbucket.org/decimalteam/go-smart-node/x/coin/types"
	feetypes "bitbucket.org/decimalteam/go-smart-node/x/fee/types"
	legacytypes "bitbucket.org/decimalteam/go-smart-node/x/legacy/types"
	nfttypes "bitbucket.org/decimalteam/go-smart-node/x/nft/types"
	swaptypes "bitbucket.org/decimalteam/go-smart-node/x/swap/types"
	validatortypes "bitbucket.org/decimalteam/go-smart-node/x/validator/types"

	"github.com/dustin/go-humanize"
	"github.com/status-im/keycard-go/hexutils"
	abci "github.com/tendermint/tendermint/abci/types"

	web3common "github.com/ethereum/go-ethereum/common"
	web3hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	web3types "github.com/ethereum/go-ethereum/core/types"
	web3 "github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/tendermint/tendermint/libs/log"

	rpc "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TxReceiptsBatchSize = 16
	RequestTimeout      = 16 * time.Second
	RequestRetryDelay   = 32 * time.Millisecond
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
	EVM               BlockEVM           `json:"evm"`
	//TODO: add EventAccumulator
	StateChanges EventAccumulator `json:"state_changes"`
}

type BlockData struct {
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

type BlockEVM struct {
	Header       *web3types.Header    `json:"header"`
	Transactions []*TransactionEVM    `json:"transactions"`
	Uncles       []*web3types.Header  `json:"uncles"`
	Receipts     []*web3types.Receipt `json:"receipts"`
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

type FailedTxLog struct {
	Log string `json:"log"`
}

var pool = map[string]bool{
	mustConvertAndEncode(authtypes.NewModuleAddress(authtypes.FeeCollectorName)):     false,
	mustConvertAndEncode(authtypes.NewModuleAddress(distrtypes.ModuleName)):          false,
	mustConvertAndEncode(authtypes.NewModuleAddress(stakingtypes.BondedPoolName)):    false,
	mustConvertAndEncode(authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)): false,
	mustConvertAndEncode(authtypes.NewModuleAddress(govtypes.ModuleName)):            false,
	mustConvertAndEncode(authtypes.NewModuleAddress(evmtypes.ModuleName)):            false,
	mustConvertAndEncode(authtypes.NewModuleAddress(cointypes.ModuleName)):           false,
	mustConvertAndEncode(authtypes.NewModuleAddress(nfttypes.ReservedPool)):          false,
	mustConvertAndEncode(authtypes.NewModuleAddress(legacytypes.LegacyCoinPool)):     false,
	mustConvertAndEncode(authtypes.NewModuleAddress(swaptypes.SwapPool)):             false,
	mustConvertAndEncode(authtypes.NewModuleAddress(validatortypes.ModuleName)):      false,
	mustConvertAndEncode(authtypes.NewModuleAddress(feetypes.BurningPool)):           false,
}

type EventAccumulator struct {
	// [address][coin_symbol]amount changes
	BalancesChanges map[string]map[string]sdkmath.Int `json:"balances_changes"`
	// [denom]vr struct
	CoinsVR       map[string]UpdateCoinVR `json:"coins_vr"`
	PayCommission []EventPayCommission    `json:"pay_commission"`
	CoinsStaked   map[string]sdkmath.Int  `json:"coin_staked"`
	// [coin_symbol]
	CoinsCreates []EventCreateCoin          `json:"-"`
	CoinUpdates  map[string]EventUpdateCoin `json:"-"`
	// replace legacy
	LegacyReown        map[string]string    `json:"legacy_reown"`
	LegacyReturnNFT    []LegacyReturnNFT    `json:"legacy_return_nft"`
	LegacyReturnWallet []LegacyReturnWallet `json:"legacy_return_wallet"`
	// multisig
	MultisigCreateWallets []MultisigCreateWallet `json:"-"`
	MultisigCreateTxs     []MultisigCreateTx     `json:"-"`
	MultisigSignTxs       []MultisigSignTx       `json:"-"`
	// nft
	//Collection    []EventUpdateCollection `json:"collection"`
	CreateToken   []EventCreateToken   `json:"-"`
	MintSubTokens []EventMintToken     `json:"-"`
	BurnSubTokens []EventBurnToken     `json:"-"`
	UpdateToken   []EventUpdateToken   `json:"-"`
	UpdateReserve []EventUpdateReserve `json:"-"`
	SendNFTs      []EventSendToken     `json:"-"`
	// swap
	ActivateChain   []EventActivateChain   `json:"-"`
	DeactivateChain []EventDeactivateChain `json:"-"`
	SwapInitialize  []EventSwapInitialize  `json:"-"`
	SwapRedeem      []EventSwapRedeem      `json:"-"`
}

type UpdateCoinVR struct {
	Volume  sdkmath.Int `json:"volume"`
	Reserve sdkmath.Int `json:"reserve"`
}

type EventUpdateCoin struct {
	LimitVolume sdkmath.Int `json:"limitVolume"`
	Avatar      string      `json:"avatar"` // identity
}

type EventCreateCoin struct {
	Denom       string      `json:"denom"`
	Title       string      `json:"title"`
	Volume      sdkmath.Int `json:"volume"`
	Reserve     sdkmath.Int `json:"reserve"`
	CRR         uint64      `json:"crr"`
	LimitVolume sdkmath.Int `json:"limitVolume"`
	Creator     string      `json:"creator"`
	Avatar      string      `json:"avatar"` // identity
	// can get from transactions
	TxHash string `json:"txHash"`
	// ? priceUSD
	// ? burn
}

func NewEventAccumulator() *EventAccumulator {
	return &EventAccumulator{
		BalancesChanges: make(map[string]map[string]sdkmath.Int),
		CoinUpdates:     make(map[string]EventUpdateCoin),
		CoinsVR:         make(map[string]UpdateCoinVR),
		CoinsStaked:     make(map[string]sdkmath.Int),
		LegacyReown:     make(map[string]string),
	}
}

type EventPayCommission struct {
	Payer  string    `json:"payer"`
	Coins  sdk.Coins `json:"coins"`
	Burnt  sdk.Coins `json:"burnt"`
	TxHash string    `json:"txHash"`
}

type LegacyReturnNFT struct {
	LegacyOwner string `json:"legacy_owner"`
	Owner       string `json:"owner"`
	Denom       string `json:"denom"`
	Creator     string `json:"creator"`
	ID          string `json:"id"`
}

type LegacyReturnWallet struct {
	LegacyOwner string `json:"legacy_owner"`
	Owner       string `json:"owner"`
	Wallet      string `json:"wallet"`
}

// decimal-models
type MultisigCreateWallet struct {
	Address   string          `json:"address"`
	Threshold uint32          `json:"threshold"`
	Creator   string          `json:"creator"`
	Owners    []MultisigOwner `json:"owners"`
}

type MultisigOwner struct {
	Address  string `json:"address"`
	Multisig string `json:"multisig"`
	Weight   uint32 `json:"weight"`
}

type MultisigCreateTx struct {
	Sender      string    `json:"sender"`
	Wallet      string    `json:"wallet"`
	Receiver    string    `json:"receiver"`
	Transaction string    `json:"transaction"`
	Coins       sdk.Coins `json:"coins"`
	TxHash      string    `json:"txHash"`
}

type MultisigSignTx struct {
	Sender        string `json:"sender"`
	Wallet        string `json:"wallet"`
	Transaction   string `json:"transaction"`
	SignerWeight  uint32 `json:"signer_weight"`
	Confirmations uint32 `json:"confirmations"`
	Confirmed     bool   `json:"confirmed"`
}

type EventCreateToken struct {
	NftID         string   `json:"nftId"`
	NftCollection string   `json:"nftCollection"`
	TokenURI      string   `json:"tokenUri"`
	Creator       string   `json:"creator"`
	StartReserve  sdk.Coin `json:"startReserve"`
	TotalReserve  sdk.Coin `json:"totalReserve"`
	AllowMint     bool     `json:"allowMint"`
	Recipient     string   `json:"recipient"`
	Quantity      uint32   `json:"quantity"`
	SubTokenIDs   []uint32 `json:"subTokenIds"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventMintToken struct {
	Creator       string   `json:"creator"`
	Recipient     string   `json:"recipient"`
	NftCollection string   `json:"nftCollection"`
	NftID         string   `json:"nftId"`
	StartReserve  sdk.Coin `json:"startReserve"`
	SubTokenIDs   []uint32 `json:"subTokenIds"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventBurnToken struct {
	Sender      string   ` json:"sender"`
	NftID       string   `json:"nftId"`
	SubTokenIDs []uint32 `json:"subTokenIds"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventUpdateToken struct {
	//Sender   string ` json:"sender"`
	NftID    string `json:"nftId"`
	TokenURI string `json:"tokenUri"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventUpdateReserve struct {
	//Sender      string   ` json:"sender"`
	NftID       string   `json:"nftId"`
	Reserve     sdk.Coin `json:"reserve"`
	Refill      sdk.Coin `json:"refill"`
	SubTokenIDs []uint32 `json:"subTokenIds"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventSendToken struct {
	Sender      string   ` json:"sender"`
	NftID       string   `json:"nftId"`
	Recipient   string   `json:"recipient"`
	SubTokenIDs []uint32 `json:"subTokenIds"`
	// from tx
	TxHash string `json:"txHash"`
}

type EventActivateChain struct {
	Sender string `json:"sender"`
	ID     uint32 `json:"id"`
	Name   string `json:"name"`
	TxHash string `json:"txHash"`
}

type EventDeactivateChain struct {
	Sender string `json:"sender"`
	ID     uint32 `json:"id"`
	TxHash string `json:"txHash"`
}

type EventSwapInitialize struct {
	Sender            string      `json:"sender"`
	DestChain         uint32      `json:"destChain"`
	FromChain         uint32      `json:"fromChain"`
	Recipient         string      `json:"recipient"`
	Amount            sdkmath.Int `json:"amount"`
	TokenDenom        string      `json:"tokenDenom"`
	TransactionNumber string      `json:"transactionNumber"`
	TxHash            string      `json:"txHash"`
}

type EventSwapRedeem struct {
	Sender            string      `json:"sender"`
	From              string      `json:"from"`
	Recipient         string      `json:"recipient"`
	Amount            sdkmath.Int `json:"amount"`
	TokenDenom        string      `json:"tokenDenom"`
	TransactionNumber string      `json:"transactionNumber"`
	DestChain         uint32      `json:"destChain"`
	FromChain         uint32      `json:"fromChain"`
	HashReedem        string      `json:"hashReedem"`
	V                 string      `json:"v"`
	R                 string      `json:"r"`
	S                 string      `json:"s"`
	TxHash            string      `json:"txHash"`
}

type processFunc func(ea *EventAccumulator, event abci.Event, txHash string) error

var eventProcessors = map[string]processFunc{
	// coins
	"decimal.coin.v1.EventCreateCoin":   processEventCreateCoin,
	"decimal.coin.v1.EventUpdateCoin":   processEventUpdateCoin,
	"decimal.coin.v1.EventUpdateCoinVR": processEventUpdateCoinVR,
	"decimal.coin.v1.EventSendCoin":     processEventSendCoin,
	"decimal.coin.v1.EventBuySellCoin":  processEventBuySellCoin,
	"decimal.coin.v1.EventBurnCoin":     processEventBurnCoin,
	"decimal.coin.v1.EventRedeemCheck":  processEventRedeemCheck,
	// fee
	"decimal.fee.v1.EventUpdateCoinPrices": processEventUpdatePrices,
	"decimal.fee.v1.EventPayCommission":    processEventPayCommission,
	// legacy
	"decimal.legacy.v1.EventReturnLegacyCoins":    processEventReturnLegacyCoins,
	"decimal.legacy.v1.EventReturnLegacySubToken": processEventReturnLegacySubToken,
	"decimal.legacy.v1.EventReturnMultisigWallet": processEventReturnMultisigWallet,
	// multisig
	"decimal.multisig.v1.EventCreateWallet":       processEventCreateWallet,
	"decimal.multisig.v1.EventCreateTransaction":  processEventCreateTransaction,
	"decimal.multisig.v1.EventSignTransaction":    processEventSignTransaction,
	"decimal.multisig.v1.EventConfirmTransaction": processEventConfirmTransaction,
	// nft
	"decimal.nft.v1.EventCreateCollection": processEventCreateCollection,
	"decimal.nft.v1.EventUpdateCollection": processEventCreateCollection,
	"decimal.nft.v1.EventCreateToken":      processEventCreateToken,
	"decimal.nft.v1.EventMintToken":        processEventMintNFT,
	"decimal.nft.v1.EventUpdateToken":      processEventUpdateToken,
	"decimal.nft.v1.EventUpdateReserve":    processEventUpdateReserve,
	"decimal.nft.v1.EventSendToken":        processEventSendNFT,
	"decimal.nft.v1.EventBurnToken":        processEventBurnNFT,
	// swap
	"decimal.swap.v1.EventActivateChain":   processEventActivateChain,
	"decimal.swap.v1.EventDeactivateChain": processEventDeactivateChain,
	"decimal.swap.v1.EventInitializeSwap":  processEventSwapInitialize,
	"decimal.swap.v1.EventRedeemSwap":      processEventSwapRedeem,
	// validator
	"decimal.validator.v1.EventDelegate":           processEventDelegate,
	"decimal.validator.v1.EventUndelegateComplete": processEventUndelegateComplete,
	"decimal.validator.v1.EventForceUndelegate":    processEventUndelegateComplete,
	"decimal.validator.v1.EventRedelegateComplete": processEventRedelegateComplete,
	"decimal.validator.v1.EventUpdateCoinsStaked":  processEventUpdateCoinsStaked,

	banktypes.EventTypeTransfer: processEventTransfer,
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
		fmt.Sprintf("Compiled block (%s)", DurationToString(time.Since(start))),
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
				fmt.Sprintf("Block is successfully sent (%s)", DurationToString(time.Since(start))),
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
					fmt.Sprintf("Fetched block (after %s)", DurationToString(time.Since(start))),
					"block", height,
				)
			} else {
				w.logger.Info(
					fmt.Sprintf("Fetched block (%s)", DurationToString(time.Since(start))),
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

func (w *Worker) fetchBlockSize(height int64, ch chan int) {

	// Request blockchain info
	result, err := w.rpcClient.BlockchainInfo(w.ctx, height, height)
	w.panicError(err)

	// Send result to the channel
	ch <- result.BlockMetas[0].BlockSize
}

func (w *Worker) fetchBlockResults(height int64, block ctypes.ResultBlock, ea *EventAccumulator, ch chan []Tx, brch chan *ctypes.ResultBlockResults) {
	var err error

	// Request block results from the node
	// NOTE: Try to retrieve results in the loop since it looks like there is some delay before results are ready to by retrieved
	var blockResults *ctypes.ResultBlockResults
	for c := 1; true; c++ {
		if c > 5 {
			w.logger.Debug(fmt.Sprintf("%d attempt to fetch block height: %d, time %s", c, height, time.Now().String()))
		}
		// Request block results
		blockResults, err = w.rpcClient.BlockResults(w.ctx, &height)
		if err == nil {
			break
		}
		// Sleep some time before next try
		time.Sleep(RequestRetryDelay)
	}

	// Prepare block results by overall processing
	var results []Tx
	for i, tx := range block.Block.Txs {
		var result Tx
		var txLog []interface{}
		txr := blockResults.TxsResults[i]

		recoveredTx, err := w.cdc.TxConfig.TxDecoder()(tx)
		w.panicError(err)

		// Parse transaction results logs
		err = json.Unmarshal([]byte(txr.Log), &txLog)
		if err != nil {
			result.Log = []interface{}{FailedTxLog{Log: txr.Log}}
		} else {
			result.Log = txLog
		}

		result.Info = w.parseTxInfo(recoveredTx)
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
				w.panicError(err)
			}
		}
	}

	// Send results to the channel
	ch <- results
	brch <- blockResults
}

func (w *Worker) fetchBlockWeb3(height int64, ch chan *web3types.Block) {

	// Request block by number
	result, err := w.web3Client.BlockByNumber(w.ctx, big.NewInt(height))
	w.panicError(err)

	// Send result to the channel
	ch <- result
}

func (w *Worker) fetchBlockTxReceiptsWeb3(block *web3types.Block, ch chan web3types.Receipts) {
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
			w.logger.Debug(fmt.Sprintf("%d attempt to fetch transaction receipts with height: %d, time %s", c, block.NumberU64(), time.Now().String()))
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
		err := w.ethRpcClient.BatchCall(requests[:])
		if err == nil {
			// Ensure all transaction receipts are retrieved
			for i := range requests {
				txHash := requests[i].Args[0].(web3common.Hash)
				if requests[i].Error != nil {
					err = requests[i].Error
					if c > 5 {
						w.logger.Error(fmt.Sprintf("Error: %v", err))
					}
					continue
				}
				if results[i].BlockNumber == nil || results[i].BlockNumber.Sign() == 0 {
					err = fmt.Errorf("got null result for tx with hash %v", txHash)
					if c > 5 {
						w.logger.Error(fmt.Sprintf("Error: %v", err))
					}
				}
			}
		}
		if err == nil {
			break
		}
		// Sleep some time before next try
		time.Sleep(RequestRetryDelay)
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

func (w *Worker) parseTxInfo(tx sdk.Tx) (txInfo TxInfo) {
	if tx == nil {
		return
	}
	for _, rawMsg := range tx.GetMsgs() {
		params := make(map[string]interface{})
		err := json.Unmarshal(w.cdc.Codec.MustMarshalJSON(rawMsg), &params)
		w.panicError(err)
		var msg TxMsg
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

func (w *Worker) parseEvents(events []abci.Event) []Event {
	var newEvents []Event
	for _, ev := range events {
		newEvent := Event{
			Type:       ev.Type,
			Attributes: []Attribute{},
		}
		for _, attr := range ev.Attributes {
			newEvent.Attributes = append(newEvent.Attributes, Attribute{
				Key:   string(attr.Key),
				Value: string(attr.Value),
			})
		}
		newEvents = append(newEvents, newEvent)
	}
	return newEvents
}

func (ea *EventAccumulator) AddEvent(event abci.Event, txHash string) error {
	procFunc, ok := eventProcessors[event.Type]
	if !ok {
		return processStub(ea, event, txHash)
	}
	return procFunc(ea, event, txHash)
}

// stub to skip internal cosmos events
func processStub(ea *EventAccumulator, event abci.Event, txHash string) error {
	return nil
}

func processEventTransfer(ea *EventAccumulator, event abci.Event, txHash string) error {
	var (
		err        error
		coins      sdk.Coins
		sender     string
		receipient string
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case banktypes.AttributeKeySender:
			sender = string(attr.Value)
		case banktypes.AttributeKeyRecipient:
			receipient = string(attr.Value)
		case sdk.AttributeKeyAmount:
			coins, err = sdk.ParseCoinsNormalized(string(attr.Value))
			if err != nil {
				return fmt.Errorf("can't parse coins: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		}
	}

	for _, coin := range coins {
		if _, ok := pool[sender]; !ok {
			ea.addBalanceChange(sender, coin.Denom, coin.Amount.Neg())
		}
		if _, ok := pool[receipient]; !ok {
			ea.addBalanceChange(receipient, coin.Denom, coin.Amount)
		}
	}

	return nil
}

// account balances
// gathered from all transactions, amount can be negative
func (ea *EventAccumulator) addBalanceChange(address string, symbol string, amount sdkmath.Int) {
	balance, ok := ea.BalancesChanges[address]
	if !ok {
		ea.BalancesChanges[address] = map[string]sdkmath.Int{symbol: amount}
		return
	}
	knownChange, ok := balance[symbol]
	if !ok {
		knownChange = sdkmath.ZeroInt()
	}
	balance[symbol] = knownChange.Add(amount)
	ea.BalancesChanges[address] = balance
}

// decimal.coin.v1.EventCreateCoin
func processEventCreateCoin(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string denom = 2;
	  string title = 3;
	  uint32 crr = 4 [ (gogoproto.customname) = "CRR" ];
	  string initial_volume = 5;
	  string initial_reserve = 6;
	  string limit_volume = 7;
	  string identity = 8;
	  string commission_create_coin = 9;
	*/
	var ecc EventCreateCoin
	var err error
	var ok bool
	var sender string
	//var commission sdk.Coin
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			sender = string(attr.Value)
			ecc.Creator = sender
		case "denom":
			ecc.Denom = string(attr.Value)
		case "title":
			ecc.Title = string(attr.Value)
		case "identity":
			ecc.Avatar = string(attr.Value)
		case "crr":
			ecc.CRR, err = strconv.ParseUint(string(attr.Value), 10, 64)
			if err != nil {
				return fmt.Errorf("can't parse crr '%s': %s", string(attr.Value), err.Error())
			}
		case "initial_volume":
			ecc.Volume, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse initial_volume '%s'", string(attr.Value))
			}
		case "initial_reserve":
			ecc.Reserve, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse initial_reserve '%s'", string(attr.Value))
			}
		case "limit_volume":
			ecc.LimitVolume, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse limit_volume '%s'", string(attr.Value))
			}
			//case "commission_create_coin":
			//	commission, err = sdk.ParseCoinNormalized(string(attr.Value))
			//	if err != nil {
			//		return fmt.Errorf("can't parse commission_create_coin '%s': %s", string(attr.Value), err.Error())
		}
	}
	ecc.TxHash = txHash
	//ea.addBalanceChange(sender, baseCoinSymbol, ecc.Reserve.Neg())
	//ea.addBalanceChange(sender, ecc.Denom, ecc.Volume)
	//ea.addBalanceChange(sender, commission.Denom, commission.Amount.Neg())

	ea.CoinsCreates = append(ea.CoinsCreates, ecc)
	return nil
}

// decimal.coin.v1.EventUpdateCoin
func processEventUpdateCoin(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string denom = 2;
	  string limit_volume = 3;
	  string identity = 4;
	*/
	var ok bool
	var euc EventUpdateCoin
	var denom string
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "denom":
			denom = string(attr.Value)
		case "identity":
			euc.Avatar = string(attr.Value)
		case "limit_volume":
			euc.LimitVolume, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse limit_volume '%s'", string(attr.Value))
			}
		}
	}
	ea.CoinUpdates[denom] = euc
	return nil
}

// decimal.coin.v1.EventUpdateCoinVR
func processEventUpdateCoinVR(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string denom = 1;
	  string volume = 2;
	  string reserve = 3;
	*/
	var ok bool
	var e UpdateCoinVR
	var denom string
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "denom":
			denom = string(attr.Value)
		case "volume":
			e.Volume, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse volume '%s'", string(attr.Value))
			}
		case "reserve":
			e.Reserve, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse reserve '%s'", string(attr.Value))
			}
		}
	}

	ea.addCoinVRChange(denom, e)
	return nil
}

// decimal.coin.v1.EventSendCoin
func processEventSendCoin(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1
	  string recipient = 2
	  string coin = 3;
	*/
	//var err error
	//var sender, recipient string
	//var coin sdk.Coin
	//for _, attr := range event.Attributes {
	//	switch string(attr.Key) {
	//	case "sender":
	//		sender = string(attr.Value)
	//	case "recipient":
	//		recipient = string(attr.Value)
	//	case "coin":
	//		coin, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse coin '%s': %s", string(attr.Value), err.Error())
	//		}
	//	}
	//}
	//
	//ea.addBalanceChange(sender, coin.Denom, coin.Amount.Neg())
	//ea.addBalanceChange(recipient, coin.Denom, coin.Amount)
	return nil
}

// decimal.coin.v1.EventBuySellCoin
func processEventBuySellCoin(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1
	  string coin_to_buy = 2;
	  string coin_to_sell = 3;
	  string amount_in_base_coin = 4;
	*/
	//var err error
	//var sender string
	//var coinToBuy, coinToSell sdk.Coin
	//for _, attr := range event.Attributes {
	//	switch string(attr.Key) {
	//	case "sender":
	//		sender = string(attr.Value)
	//	case "coin_to_buy":
	//		coinToBuy, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse coin '%s': %s", string(attr.Value), err.Error())
	//		}
	//	case "coin_to_sell":
	//		coinToSell, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse coin '%s': %s", string(attr.Value), err.Error())
	//		}
	//	}
	//}
	//
	//ea.addBalanceChange(sender, coinToBuy.Denom, coinToBuy.Amount)
	//ea.addBalanceChange(sender, coinToSell.Denom, coinToSell.Amount.Neg())
	return nil
}

// decimal.coin.v1.EventBurnCoin
func processEventBurnCoin(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1
	  string coin = 2;
	*/
	//var err error
	//var sender string
	//var coinToBurn sdk.Coin
	//for _, attr := range event.Attributes {
	//	switch string(attr.Key) {
	//	case "sender":
	//		sender = string(attr.Value)
	//	case "coin":
	//		coinToBurn, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse coin '%s': %s", string(attr.Value), err.Error())
	//		}
	//	}
	//}

	//ea.addBalanceChange(sender, coinToBurn.Denom, coinToBurn.Amount.Neg())
	return nil
}

// decimal.coin.v1.EventRedeemCheck
func processEventRedeemCheck(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1
	  string issuer = 2
	  string coin = 3;
	  string nonce = 4;
	  string due_block = 5;
	  string commission_redeem_check = 6;
	*/
	//var err error
	//var sender, issuer string
	//var coin, commission sdk.Coin
	//for _, attr := range event.Attributes {
	//	if string(attr.Key) == "sender" {
	//		sender = string(attr.Value)
	//	}
	//	if string(attr.Key) == "issuer" {
	//		issuer = string(attr.Value)
	//	}
	//	if string(attr.Key) == "coin" {
	//		coin, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse coin '%s': %s", string(attr.Value), err.Error())
	//		}
	//	}
	//	if string(attr.Key) == "commission_redeem_check" {
	//		commission, err = sdk.ParseCoinNormalized(string(attr.Value))
	//		if err != nil {
	//			return fmt.Errorf("can't parse commission_redeem_check '%s': %s", string(attr.Value), err.Error())
	//		}
	//	}
	//}

	//ea.addBalanceChange(sender, coin.Denom, coin.Amount)
	//ea.addBalanceChange(issuer, coin.Denom, coin.Amount.Neg())
	//ea.addBalanceChange(issuer, commission.Denom, commission.Amount.Neg())
	return nil
}

func mustConvertAndEncode(address sdk.AccAddress) string {
	res, err := bech32.ConvertAndEncode(cmdcfg.Bech32PrefixAccAddr, address)
	if err != nil {
		panic(err)
	}

	return res
}

func processEventUpdatePrices(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string oracle = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		repeated CoinPrice prices = 2 [ (gogoproto.nullable) = false ];
	*/

	// TODO this event need handle?

	return nil
}

// decimal.fee.v1.EventPayCommission
func processEventPayCommission(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string payer = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  repeated cosmos.base.v1beta1.Coin coins = 2
	  [ (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins", (gogoproto.nullable) = false ];
	*/
	var (
		err   error
		coins sdk.Coins
		burnt sdk.Coins
		e     EventPayCommission
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "payer":
			e.Payer = string(attr.Value)
		case "coins":
			err = json.Unmarshal(attr.Value, &coins)
			if err != nil {
				return fmt.Errorf("can't unmarshal coins: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		case "burnt":
			err = json.Unmarshal(attr.Value, &burnt)
			if err != nil {
				return fmt.Errorf("can't unmarshal burnt coins: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		}
	}

	e.Coins = coins
	e.TxHash = txHash
	e.Burnt = burnt
	ea.PayCommission = append(ea.PayCommission, e)
	return nil

}

// decimal.legacy.v1.EventReturnLegacyCoins
func processEventReturnLegacyCoins(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string legacy_owner = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string owner = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  repeated cosmos.base.v1beta1.Coin coins = 3 [
	    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
	    (gogoproto.nullable) = false
	  ];
	*/
	var err error
	var oldAddress, newAddress string
	var coins sdk.Coins
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "legacy_owner":
			oldAddress = string(attr.Value)
		case "owner":
			newAddress = string(attr.Value)
		case "coins":
			err = json.Unmarshal(attr.Value, &coins)
			if err != nil {
				return fmt.Errorf("can't parse coins '%s': %s", string(attr.Value), err.Error())
			}
		}
	}
	//for _, coin := range coins {
	//ea.addBalanceChange(newAddress, coin.Denom, coin.Amount)
	//}
	ea.LegacyReown[oldAddress] = newAddress
	return nil

}

// decimal.legacy.v1.EventReturnLegacySubToken
func processEventReturnLegacySubToken(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string legacy_owner = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string owner = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string denom = 3;
	  string id = 4 [ (gogoproto.customname) = "ID" ];
	  repeated uint32 sub_token_ids = 5 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/
	var ret LegacyReturnNFT
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "legacy_owner":
			ret.LegacyOwner = string(attr.Value)
		case "owner":
			ret.Owner = string(attr.Value)
		case "denom":
			ret.Denom = string(attr.Value)
		case "id":
			ret.ID = string(attr.Value)
		}
	}
	ea.LegacyReturnNFT = append(ea.LegacyReturnNFT, ret)
	return nil

}

// decimal.legacy.v1.EventReturnMultisigWallet
func processEventReturnMultisigWallet(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string legacy_owner = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string owner = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string wallet = 3 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	*/
	var ret LegacyReturnWallet
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "legacy_owner":
			ret.LegacyOwner = string(attr.Value)
		case "owner":
			ret.Owner = string(attr.Value)
		case "wallet":
			ret.Wallet = string(attr.Value)
		}
	}
	ea.LegacyReturnWallet = append(ea.LegacyReturnWallet, ret)
	return nil

}
