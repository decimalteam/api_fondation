package api_fondation

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

type processFunc func(ea *EventAccumulator, event abci.Event, txHash string) error

