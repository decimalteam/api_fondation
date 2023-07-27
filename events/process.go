package events

import (
	"api_fondation/types"
	sdkmath "cosmossdk.io/math"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
)

var pool = map[string]bool{
	mustConvertAndEncode(authtypes.NewModuleAddress(authtypes.FeeCollectorName)):      false,
	mustConvertAndEncode(authtypes.NewModuleAddress(distrtypes.ModuleName)):           false,
	mustConvertAndEncode(authtypes.NewModuleAddress(stakingtypes.BondedPoolName)):     false,
	mustConvertAndEncode(authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName)):  false,
	mustConvertAndEncode(authtypes.NewModuleAddress(govtypes.ModuleName)):             false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.EvmtypesModuleName)):        false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.CointypesModuleName)):       false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.NfttypesReservedPool)):      false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.LegacytypesLegacyCoinPool)): false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.SwaptypesSwapPool)):         false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.ValidatortypesModuleName)):  false,
	mustConvertAndEncode(authtypes.NewModuleAddress(types.FeetypesBurningPool)):       false,
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
	res, err := bech32.ConvertAndEncode(types.Bech32PrefixAccAddr, address)
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

// decimal.multisig.v1.EventCreateWallet
func processEventCreateWallet(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string wallet = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  repeated string owners = 3;
	  repeated uint32 weights = 4;
	  uint32 threshold = 5;
	*/
	e := MultisigCreateWallet{}
	var owners []string
	var weights []uint32
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Creator = string(attr.Value)
		case "wallet":
			e.Address = string(attr.Value)
		case "threshold":
			thr, err := strconv.ParseUint(string(attr.Value), 10, 64)
			if err != nil {
				return fmt.Errorf("can't parse threshold '%s': %s", string(attr.Value), err.Error())
			}
			e.Threshold = uint32(thr)
		case "owners":
			err := json.Unmarshal(attr.Value, &owners)
			if err != nil {
				return fmt.Errorf("can't unmarshal owners: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		case "weights":
			err := json.Unmarshal(attr.Value, &weights)
			if err != nil {
				return fmt.Errorf("can't unmarshal weights: %s", err.Error())
			}
		}
	}
	for i, owner := range owners {
		e.Owners = append(e.Owners, MultisigOwner{
			Address:  owner,
			Multisig: e.Address,
			Weight:   weights[i],
		})
	}

	ea.MultisigCreateWallets = append(ea.MultisigCreateWallets, e)
	return nil
}

// decimal.multisig.v1.EventCreateTransaction
func processEventCreateTransaction(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		string wallet = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		string receiver = 3 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		string coins = 4;
		string transaction = 5;
	*/
	e := MultisigCreateTx{}
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "wallet":
			e.Wallet = string(attr.Value)
		case "receiver":
			e.Receiver = string(attr.Value)
		case "coins":
			err := json.Unmarshal(attr.Value, &e.Coins)
			if err != nil {
				return fmt.Errorf("can't unmarshal coins: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		case "transaction":
			e.Transaction = string(attr.Value)
		}
	}

	e.TxHash = txHash
	ea.MultisigCreateTxs = append(ea.MultisigCreateTxs, e)
	return nil
}

// decimal.multisig.v1.EventSignTransaction
func processEventSignTransaction(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string wallet = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string transaction = 3;
	  uint32 signer_weight = 4;
	  uint32 confirmations = 5;
	  bool confirmed = 6;
	*/
	e := MultisigSignTx{}
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "wallet":
			e.Wallet = string(attr.Value)
		case "transaction":
			e.Transaction = string(attr.Value)
		case "signer_weight":
			err := json.Unmarshal(attr.Value, &e.SignerWeight)
			if err != nil {
				return fmt.Errorf("can't unmarshal signer_weight: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		case "confirmations":
			err := json.Unmarshal(attr.Value, &e.Confirmations)
			if err != nil {
				return fmt.Errorf("can't unmarshal confirmations: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		case "confirmed":
			err := json.Unmarshal(attr.Value, &e.Confirmed)
			if err != nil {
				return fmt.Errorf("can't unmarshal confirmed: %s, value: '%s'", err.Error(), string(attr.Value))
			}
		}
	}
	ea.MultisigSignTxs = append(ea.MultisigSignTxs, e)
	return nil
}

// decimal.multisig.v1.EventConfirmTransaction
func processEventConfirmTransaction(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string wallet = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string receiver = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string transaction = 3;
	  repeated cosmos.base.v1beta1.Coin coins = 4
	      [ (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins", (gogoproto.nullable) = false ];
	*/
	//var wallet, receiver string
	//var coins = sdk.NewCoins()
	//for _, attr := range event.Attributes {
	//	switch string(attr.Key) {
	//	case "wallet":
	//		wallet = string(attr.Value)
	//	case "receiver":
	//		receiver = string(attr.Value)
	//	case "coins":
	//		err := json.Unmarshal(attr.Value, &coins)
	//		if err != nil {
	//			return fmt.Errorf("can't unmarshal coins: %s, value: '%s'", err.Error(), string(attr.Value))
	//		}
	//	}
	//for _, coin := range coins {
	//	ea.addBalanceChange(wallet, coin.Denom, coin.Amount.Neg())
	//	ea.addBalanceChange(receiver, coin.Denom, coin.Amount)
	//}
	return nil
}

func processEventCreateCollection(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string denom = 2;
	  uint32 supply = 3;
	*/
	//var err error
	//var e EventUpdateCollection
	//for _, attr := range event.Attributes {
	//	switch string(attr.Key) {
	//	case "creator":
	//		e.Creator = string(attr.Value)
	//	case "denom":
	//		e.Denom = string(attr.Value)
	//	case "supply":
	//		e.Supply = binary.LittleEndian.Uint32(attr.Value)
	//	}
	//}
	//e.TxHash = txHash
	//
	//ea.Collection = append(ea.Collection, e)
	return nil
}

func processEventCreateToken(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		string denom = 2;
		string id = 3 [ (gogoproto.customname) = "ID" ];
		string uri = 4 [ (gogoproto.customname) = "URI" ];
		bool allowMint = 5;
		string reserve = 6;
		string recipient = 7;
		repeated uint32 subTokenIds = 8 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/
	var err error
	var e EventCreateToken
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "creator":
			e.Creator = string(attr.Value)
		case "denom":
			e.NftCollection = string(attr.Value)
		case "id":
			e.NftID = string(attr.Value)
		case "uri":
			e.TokenURI = string(attr.Value)
		case "allowMint":
			e.AllowMint = false
			if string(attr.Value) == "true" {
				e.AllowMint = true
			}
		case "reserve":
			e.StartReserve, err = sdk.ParseCoinNormalized(string(attr.Value))
			if err != nil {
				return fmt.Errorf("can't parse reserve '%s': %s", string(attr.Value), err.Error())
			}
		case "subTokenIds":
			err := json.Unmarshal(attr.Value, &e.SubTokenIDs)
			if err != nil {
				return fmt.Errorf("can't unmarshal subTokenIds: %s", err.Error())
			}
		}
	}

	// TODO возможно стоит убрать поля которые есть в mint из create token
	e.TxHash = txHash

	e.TotalReserve = sdk.NewCoin(e.StartReserve.Denom, e.StartReserve.Amount.Mul(sdk.NewInt(int64(len(e.SubTokenIDs)))))
	e.Quantity = uint32(len(e.SubTokenIDs))
	//ea.addBalanceChange(e.Creator, e.TotalReserve.Denom, e.TotalReserve.Amount.Neg())
	//ea.addBalanceChange(reservedPool.String(), e.TotalReserve.Denom, e.TotalReserve.Amount)
	ea.addMintSubTokens(EventMintToken{
		Creator:       e.Creator,
		Recipient:     e.Recipient,
		NftCollection: e.NftCollection,
		NftID:         e.NftID,
		StartReserve:  e.StartReserve,
		SubTokenIDs:   e.SubTokenIDs,
		TxHash:        txHash,
	})

	ea.CreateToken = append(ea.CreateToken, e)
	return nil
}

func processEventMintNFT(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string creator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string denom = 2;
	  string id = 3 [ (gogoproto.customname) = "ID" ];
	  string reserve = 4;
	  string recipient = 5;
	  repeated uint32 sub_token_ids = 6 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/
	var err error
	var e EventMintToken
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "creator":
			e.Creator = string(attr.Value)
		case "denom":
			e.NftCollection = string(attr.Value)
		case "id":
			e.NftID = string(attr.Value)
		case "reserve":
			e.StartReserve, err = sdk.ParseCoinNormalized(string(attr.Value))
			if err != nil {
				return fmt.Errorf("can't parse reserve '%s': %s", string(attr.Value), err.Error())
			}
		case "sub_token_ids":
			err := json.Unmarshal(attr.Value, &e.SubTokenIDs)
			if err != nil {
				return fmt.Errorf("can't unmarshal subTokenIds: %s", err.Error())
			}
		}
	}
	e.TxHash = txHash

	//totalReserve := sdk.NewCoin(e.StartReserve.Denom, e.StartReserve.Amount.Mul(sdk.NewInt(int64(len(e.SubTokenIDs)))))
	//ea.addBalanceChange(e.Creator, totalReserve.Denom, totalReserve.Amount.Neg())
	//ea.addBalanceChange(reservedPool.String(), totalReserve.Denom, totalReserve.Amount)

	ea.addMintSubTokens(e)

	return nil
}

func processEventUpdateToken(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string id = 2 [ (gogoproto.customname) = "ID" ];
	  string uri = 3 [ (gogoproto.customname) = "URI" ];
	*/

	//var err error
	var e EventUpdateToken
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "id":
			e.NftID = string(attr.Value)
		case "uri":
			e.TokenURI = string(attr.Value)
		}
	}

	ea.UpdateToken = append(ea.UpdateToken, e)
	return nil
}

func processEventUpdateReserve(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string id = 2 [ (gogoproto.customname) = "ID" ];
	  string reserve = 3; // coin that defines new reserve for all updating NFT-subtokens
	  string refill = 4;  // coin that was added in total per transaction for all NFT sub-tokens
	  repeated uint32 sub_token_ids = 5 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/
	var (
		//sender string
		err error
		e   EventUpdateReserve
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			//sender = string(attr.Value)
		case "id":
			e.NftID = string(attr.Value)
		case "reserve":
			e.Reserve, err = sdk.ParseCoinNormalized(string(attr.Value))
			if err != nil {
				return fmt.Errorf("can't parse reserve '%s': %s", string(attr.Value), err.Error())
			}
		case "refill":
			e.Refill, err = sdk.ParseCoinNormalized(string(attr.Value))
			if err != nil {
				return fmt.Errorf("can't parse reserve '%s': %s", string(attr.Value), err.Error())
			}
		case "sub_token_ids":
			err := json.Unmarshal(attr.Value, &e.SubTokenIDs)
			if err != nil {
				return fmt.Errorf("can't unmarshal subTokenIds: %s", err.Error())
			}
		}
	}

	//ea.addBalanceChange(sender, e.Refill.Denom, e.Refill.Amount.Neg())
	//ea.addBalanceChange(reservedPool.String(), e.Refill.Denom, e.Refill.Amount)
	ea.UpdateReserve = append(ea.UpdateReserve, e)

	return nil
}

func processEventSendNFT(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string id = 2 [ (gogoproto.customname) = "ID" ];
	  string recipient = 3 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  repeated uint32 sub_token_ids = 4 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/

	var e EventSendToken
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "id":
			e.NftID = string(attr.Value)
		case "recipient":
			e.Recipient = string(attr.Value)
		case "sub_token_ids":
			err := json.Unmarshal(attr.Value, &e.SubTokenIDs)
			if err != nil {
				return fmt.Errorf("can't unmarshal subTokenIds: %s", err.Error())
			}
		}
	}

	ea.SendNFTs = append(ea.SendNFTs, e)
	return nil
}

func processEventBurnNFT(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string sender = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string id = 2 [ (gogoproto.customname) = "ID" ];
	  string return = 3;  // coin that was returned in total per transaction for all NFT sub-tokens
	  repeated uint32 sub_token_ids = 4 [ (gogoproto.customname) = "SubTokenIDs" ];
	*/
	var (
		//err         error
		//returnCoins sdk.Coin
		e EventBurnToken
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "id":
			e.NftID = string(attr.Value)
		case "return":
			//returnCoins, err = sdk.ParseCoinNormalized(string(attr.Value))
			//if err != nil {
			//	return fmt.Errorf("can't parse reserve '%s': %s", string(attr.Value), err.Error())
			//}
		case "sub_token_ids":
			err := json.Unmarshal(attr.Value, &e.SubTokenIDs)
			if err != nil {
				return fmt.Errorf("can't unmarshal subTokenIds: %s", err.Error())
			}
		}
	}
	e.TxHash = txHash

	//ea.addBalanceChange(e.Sender, returnCoins.Denom, returnCoins.Amount)
	//ea.addBalanceChange(reservedPool.String(), returnCoins.Denom, returnCoins.Amount.Neg())
	ea.addBurnSubTokens(e)

	return nil
}

func processEventActivateChain(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string sender
		uint32 id = 1 [ (gogoproto.customname) = "ID" ];
		string name = 2;
	*/

	var e EventActivateChain
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "id":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.ID = uint32(id)
		case "name":
			e.Name = string(attr.Value)
		}
	}

	e.TxHash = txHash
	ea.ActivateChain = append(ea.ActivateChain, e)
	return nil
}

func processEventDeactivateChain(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string sender;
		uint32 id = 1 [ (gogoproto.customname) = "ID" ];
	*/

	var e EventDeactivateChain
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "id":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.ID = uint32(id)
		}
	}

	e.TxHash = txHash
	ea.DeactivateChain = append(ea.DeactivateChain, e)
	return nil
}

func processEventSwapInitialize(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	   string sender = 1;
	   string recipient = 5;
	   string amount = 6;
	   string token_symbol = 8;
	   string transaction_number = 7;
	   uint32 from_chain = 3;
	   uint32 dest_chain = 4;
	*/

	var (
		e EventSwapInitialize
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "recipient":
			e.Recipient = string(attr.Value)
		case "amount":
			amountStr := string(attr.Value)
			amount, ok := sdkmath.NewIntFromString(amountStr)
			if !ok {
				return fmt.Errorf("failed to parse amount: %s", amountStr)
			}
			e.Amount = amount
		case "token_symbol":
			e.Sender = string(attr.Value)
		case "transaction_number":
			e.TransactionNumber = string(attr.Value)
		case "dest_chain":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.DestChain = uint32(id)
		case "from_chain":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.FromChain = uint32(id)
		}
	}

	e.TxHash = txHash
	//ea.addBalanceChange(e.Sender, e.TokenDenom, e.Amount.Neg())
	//ea.addBalanceChange(swaptypes.SwapPool, e.TokenDenom, e.Amount)
	ea.SwapInitialize = append(ea.SwapInitialize, e)
	return nil
}

func processEventSwapRedeem(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	   string sender = 1;
	   string from = 2;
	   string recipient =3;
	   string amount = 4;
	   string token_symbol = 5;
	   string transaction_number = 6;
	   uint32 from_chain = 7;
	   uint32 dest_chain = 8;
	   string hashRedeem = 9;
	   string v = 10;
	   string r = 11;
	   string s = 12;
	*/
	var (
		e EventSwapRedeem
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "sender":
			e.Sender = string(attr.Value)
		case "from":
			e.From = string(attr.Value)
		case "recipient":
			e.Recipient = string(attr.Value)
		case "amount":
			amountStr := string(attr.Value)
			amount, ok := sdkmath.NewIntFromString(amountStr)
			if !ok {
				return fmt.Errorf("failed to parse amount: %s", amountStr)
			}
			e.Amount = amount
		case "token_symbol":
			e.Sender = string(attr.Value)
		case "transaction_number":
			e.TransactionNumber = string(attr.Value)
		case "dest_chain":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.DestChain = uint32(id)
		case "src_chain":
			strId := string(attr.Value)
			id, err := strconv.ParseUint(strId, 10, 32)
			if err != nil {
				return fmt.Errorf("can't parse chain id '%s': %s", strId, err.Error())
			}
			e.FromChain = uint32(id)
		case "v":
			e.V = string(attr.Value)
		case "r":
			e.R = string(attr.Value)
		case "s":
			e.S = string(attr.Value)
		case "hash_redeem":
			e.HashReedem = string(attr.Value)
		}
	}

	e.TxHash = txHash
	//ea.addBalanceChange(e.Recipient, e.TokenDenom, e.Amount)
	//ea.addBalanceChange(swaptypes.SwapPool, e.TokenDenom, e.Amount.Neg())
	ea.SwapRedeem = append(ea.SwapRedeem, e)
	return nil
}

func processEventDelegate(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string delegator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  string validator = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
	  Stake stake = 3 [ (gogoproto.nullable) = false ];
	  string amount_base = 4 [
	    (cosmos_proto.scalar) = "cosmos.Int",
	    (gogoproto.customtype) = "cosmossdk.io/math.Int",
	    (gogoproto.nullable) = false
	  ];
	*/

	var (
		e EventDelegate
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "delegator":
			e.Delegator = string(attr.Value)
		case "validator":
			e.Validator = string(attr.Value)
		case "stake":
			var stake Stake

			err := json.Unmarshal(attr.Value, &stake)
			if err != nil {
				panic(err)
			}
			e.Stake = stake
		}
	}

	if _, ok := pool[e.Delegator]; !ok {
		if e.Stake.Type == "STAKE_TYPE_COIN" {
			ea.addBalanceChange(e.Delegator, e.Stake.Stake.Denom, e.Stake.Stake.Amount.Neg())
		}
	}

	return nil
}

func processEventUndelegateComplete(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string delegator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		string validator = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		Stake stake = 3 [ (gogoproto.nullable) = false ];
	*/

	var (
		e EventUndelegateComplete
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "delegator":
			e.Delegator = string(attr.Value)
		case "validator":
			e.Validator = string(attr.Value)
		case "stake":
			var stake Stake

			err := json.Unmarshal(attr.Value, &stake)
			if err != nil {
				panic(err)
			}
			e.Stake = stake
		}
	}

	if _, ok := pool[e.Delegator]; !ok {
		if e.Stake.Type == "STAKE_TYPE_COIN" {
			ea.addBalanceChange(e.Delegator, e.Stake.Stake.Denom, e.Stake.Stake.Amount)
		}
	}

	return nil
}

func processEventRedelegateComplete(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
		string delegator = 1 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		  string validator_src = 2 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		  string validator_dst = 3 [ (cosmos_proto.scalar) = "cosmos.AddressString" ];
		  Stake stake = 4 [ (gogoproto.nullable) = false ];
	*/

	var (
		e EventRedelegateComplete
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "delegator":
			e.Delegator = string(attr.Value)
		case "validator_src":
			e.ValidatorSrc = string(attr.Value)
		case "validator_dst":
			e.ValidatorDst = string(attr.Value)
		case "stake":
			var stake Stake

			err := json.Unmarshal(attr.Value, &stake)
			if err != nil {
				panic(err)
			}
			e.Stake = stake
		}
	}

	//if _, ok := pool[e.Delegator]; !ok {
	//	ea.addBalanceChange(e.Delegator, e.Stake.Stake.Denom, e.Stake.Stake.Amount)
	//}

	return nil
}

func processEventUpdateCoinsStaked(ea *EventAccumulator, event abci.Event, txHash string) error {
	/*
	  string denom = 1;
	  string total_amount = 2 [
	    (cosmos_proto.scalar) = "cosmos.Int",
	    (gogoproto.customtype) = "cosmossdk.io/math.Int",
	    (gogoproto.nullable) = false
	  ];
	*/

	var (
		e  EventUpdateCoinsStaked
		ok bool
	)
	for _, attr := range event.Attributes {
		switch string(attr.Key) {
		case "denom":
			e.denom = string(attr.Value)
		case "total_amount":
			e.amount, ok = sdk.NewIntFromString(string(attr.Value))
			if !ok {
				return fmt.Errorf("can't parse total_amount '%s'", string(attr.Value))
			}
		}
	}

	ea.addCoinsStaked(e)
	return nil
}
