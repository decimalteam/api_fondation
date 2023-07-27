package events

import (
	"api_fondation"
	sdkmath "cosmossdk.io/math"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewEventAccumulator() *main.EventAccumulator {
	return &main.EventAccumulator{
		BalancesChanges: make(map[string]map[string]sdkmath.Int),
		CoinUpdates:     make(map[string]main.EventUpdateCoin),
		CoinsVR:         make(map[string]main.UpdateCoinVR),
		CoinsStaked:     make(map[string]sdkmath.Int),
		LegacyReown:     make(map[string]string),
	}
}

// account balances
// gathered from all transactions, amount can be negative
func (ea *main.EventAccumulator) addBalanceChange(address string, symbol string, amount sdkmath.Int) {
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
func (ea *main.EventAccumulator) addCoinVRChange(symbol string, vr main.UpdateCoinVR) {
	ea.CoinsVR[symbol] = vr
}

func (ea *main.EventAccumulator) addMintSubTokens(e main.EventMintToken) {
	ea.MintSubTokens = append(ea.MintSubTokens, e)
}

func (ea *main.EventAccumulator) addBurnSubTokens(e main.EventBurnToken) {
	ea.BurnSubTokens = append(ea.BurnSubTokens, e)
}

func (ea *main.EventAccumulator) addCoinsStaked(e main.EventUpdateCoinsStaked) {
	ea.CoinsStaked[e.denom] = e.amount
}

func (ea *main.EventAccumulator) AddEvent(event abci.Event, txHash string) error {
	procFunc, ok := main.eventProcessors[event.Type]
	if !ok {
		return main.processStub(ea, event, txHash)
	}
	return procFunc(ea, event, txHash)
}
