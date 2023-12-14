package worker

import (
	"bitbucket.org/decimalteam/go-smart-node/x/coin"
	"bitbucket.org/decimalteam/go-smart-node/x/legacy"
	"bitbucket.org/decimalteam/go-smart-node/x/multisig"
	"bitbucket.org/decimalteam/go-smart-node/x/nft"
	"bitbucket.org/decimalteam/go-smart-node/x/swap"
	"bitbucket.org/decimalteam/go-smart-node/x/validator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	fee "github.com/cosmos/ibc-go/v5/modules/apps/29-fee"
	ibc "github.com/cosmos/ibc-go/v5/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v5/modules/core/02-client/client"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/evm"
)

func GetModuleBasics() module.BasicManager {
	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	ModuleBasics := module.NewBasicManager(
		// SDK
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		//staking.AppModuleBasic{},
		//distr.AppModuleBasic{},
		gov.NewAppModuleBasic([]govclient.ProposalHandler{
			// SDK proposal handlers
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.LegacyProposalHandler,
			upgradeclient.LegacyCancelProposalHandler,
			// IBC proposal handlers
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			// Decimal proposal handlers
			// TODO: ...
		}),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		//slashing.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		//evidence.AppModuleBasic{},
		// IBC
		ibc.AppModuleBasic{},
		// Ethermint
		evm.AppModuleBasic{},
		// Decimal
		coin.AppModuleBasic{},
		legacy.AppModuleBasic{},
		nft.AppModuleBasic{},
		multisig.AppModuleBasic{},
		swap.AppModuleBasic{},
		fee.AppModuleBasic{},
		validator.AppModuleBasic{},
	)

	return ModuleBasics
}

const (
	// Bech32Prefix defines the Bech32 prefix used for EthAccounts
	Bech32Prefix = "d0"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32Prefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

const (
	// BaseDenom defines to the default denomination used in Decimal (staking, EVM, governance, etc.)
	// TODO: Load it from
	BaseDenom = "del"
)

// SetBech32Prefixes sets the global prefixes to be used when serializing addresses and public keys to Bech32 strings.
func SetBech32Prefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetPurpose(sdk.Purpose)
	config.SetCoinType(ethermint.Bip44CoinType)
}

// RegisterBaseDenom registers the base denomination to the SDK.
func RegisterBaseDenom() {
	if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, ethermint.BaseDenomUnit)); err != nil {
		panic(err)
	}
}
