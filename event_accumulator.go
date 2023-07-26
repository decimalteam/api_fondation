package api_fondation

import (
	sdkmath "cosmossdk.io/math"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewEventAccumulator() *EventAccumulator {
	return &EventAccumulator{
		BalancesChanges: make(map[string]map[string]sdkmath.Int),
		CoinUpdates:     make(map[string]EventUpdateCoin),
		CoinsVR:         make(map[string]UpdateCoinVR),
		CoinsStaked:     make(map[string]sdkmath.Int),
		LegacyReown:     make(map[string]string),
	}
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
