package parser

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/types"
)

type IndexData struct {
	Height  int64         `json:"height"`
	Data    *cosmos.Block `json:"data"`
	EvmData *evm.BlockEVM `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(height int64) (*IndexData, error) {
	var dataBlock *IndexData

	url := fmt.Sprintf("%s/getBlock?height=%d", p.IndexNode, height)
	bytes := client.GetRequest(url)

	err := json.Unmarshal(bytes, &dataBlock)
	if err != nil {
		p.Logger.Errorf("unmarshal data from indexer error: %v", err)
		return nil, err
	}

	return dataBlock, nil
}

func (p *Parser) getIndexerBlocksFromToHeight(heightFrom int64, heightTo int64) []types.BlockData {

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

func (p *Parser) getBlocksFromIndexer(heightFrom int64, heightTo int64) []types.BlockData {
	//var res *types.BlockData

	url := fmt.Sprintf("%s/getBlocks?height_from=%d&height_to=%d", p.IndexNode, heightFrom, heightTo)
	bytes := client.GetRequest(url)

	var dataBlocks []IndexData
	var parseBlocks []types.BlockData

	err := json.Unmarshal(bytes, &dataBlocks)
	p.Logger.Errorf("data blocks unmarshal error: %v", err)

	for _, val := range dataBlocks {
		parseBlocks = append(parseBlocks, types.BlockData{
			CosmosBlock: val.Data,
			EvmBlock:    val.EvmData,
		})
	}

	return parseBlocks
}
