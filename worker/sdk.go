package worker

import (
	"bitbucket.org/decimalteam/go-smart-node/x/coin"
	"bitbucket.org/decimalteam/go-smart-node/x/legacy"
	"bitbucket.org/decimalteam/go-smart-node/x/multisig"
	"bitbucket.org/decimalteam/go-smart-node/x/swap"
	"bitbucket.org/decimalteam/go-smart-node/x/validator"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	nft "github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	fee "github.com/cosmos/ibc-go/v5/modules/apps/29-fee"
	ibc "github.com/cosmos/ibc-go/v5/modules/core"
	"github.com/evmos/ethermint/x/evm"

	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"

	ibcclientclient "github.com/cosmos/ibc-go/v5/modules/core/02-client/client"
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
