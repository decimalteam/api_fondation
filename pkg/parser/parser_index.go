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

func (p *Parser) getBlocksFromToHeight(heightFrom int64, heightTo int64) []types.BlockData {

	url := fmt.Sprintf("%s/getBlocks?height_from=%d&height_to=%d", p.IndexNode, heightFrom, heightTo)
	bytes := client.GetRequest(url)

	var dataBlocks []IndexData
	var parseBlocks []types.BlockData

	_ = json.Unmarshal(bytes, &dataBlocks)

	for _, val := range dataBlocks {
		parseBlocks = append(parseBlocks, types.BlockData{
			CosmosBlock: val.Data,
			EvmBlock:    val.EvmData,
		})
	}

	return parseBlocks
}
