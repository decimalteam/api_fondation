package api_fondation

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	web3common "github.com/ethereum/go-ethereum/common"
	web3hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	web3types "github.com/ethereum/go-ethereum/core/types"
	web3 "github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	abci "github.com/tendermint/tendermint/abci/types"
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

type EventDelegate struct {
	Delegator string
	Validator string
	Stake     Stake
}

type EventUndelegateComplete struct {
	Delegator string
	Validator string
	Stake     Stake
}

type EventRedelegateComplete struct {
	Delegator    string
	ValidatorSrc string
	ValidatorDst string
	Stake        Stake
}

type EventUpdateCoinsStaked struct {
	denom  string
	amount sdkmath.Int
}

type Stake struct {
	Stake sdk.Coin `json:"stake"`
	Type  string   `json:"type"`
}

type processFunc func(ea *EventAccumulator, event abci.Event, txHash string) error
