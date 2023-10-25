package parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/types"
	"encoding/json"
	"fmt"
)

type IndexData struct {
	Height  int64         `json:"height"`
	Data    *cosmos.Block `json:"data"`
	EvmData *evm.BlockEVM `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(height int64) {
	//var res *types.BlockData

	url := fmt.Sprintf("%s/getBlock?height=%d", p.IndexNode, height)
	bytes := client.GetRequest(url)

	var dataBlock IndexData

	_ = json.Unmarshal(bytes, &dataBlock)

	p.NewBlockData = &types.BlockData{
		CosmosBlock: dataBlock.Data,
		EvmBlock:    dataBlock.EvmData,
	}
}
