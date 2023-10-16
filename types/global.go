package types

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
)

type LastBlockData struct {
	Height    int
	Timestamp int64
}

type BlockData struct {
	CosmosBlock *cosmos.Block
	EvmBlock    *evm.BlockEVM
}
