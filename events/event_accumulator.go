package events

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

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

func NewEventAccumulator() *EventAccumulator {
	return &EventAccumulator{
		BalancesChanges: make(map[string]map[string]sdkmath.Int),
		CoinUpdates:     make(map[string]EventUpdateCoin),
		CoinsVR:         make(map[string]UpdateCoinVR),
		CoinsStaked:     make(map[string]sdkmath.Int),
		LegacyReown:     make(map[string]string),
	}
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

// custom coin reserve or volume update
func (ea *EventAccumulator) addCoinVRChange(symbol string, vr UpdateCoinVR) {
	ea.CoinsVR[symbol] = vr
}

func (ea *EventAccumulator) addMintSubTokens(e EventMintToken) {
	ea.MintSubTokens = append(ea.MintSubTokens, e)
}

func (ea *EventAccumulator) addBurnSubTokens(e EventBurnToken) {
	ea.BurnSubTokens = append(ea.BurnSubTokens, e)
}

func (ea *EventAccumulator) addCoinsStaked(e EventUpdateCoinsStaked) {
	ea.CoinsStaked[e.denom] = e.amount
}

func (ea *EventAccumulator) AddEvent(event abci.Event, txHash string) error {
	procFunc, ok := eventProcessors[event.Type]
	if !ok {
		return processStub(ea, event, txHash)
	}
	return procFunc(ea, event, txHash)
}
