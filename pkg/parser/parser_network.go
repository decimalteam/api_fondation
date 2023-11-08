package parser

import (
	"bitbucket.org/decimalteam/api_fondation/types"
	"bitbucket.org/decimalteam/api_fondation/worker"
)

type NetworkData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromNetwork(height int64) {
	p.NewBlockData = worker.GetBlockResult(height)
}

func (p *Parser) getEvmBlock(height int64) {
	p.NewBlockData = worker.GetEvmBlock(height)
}

func (p *Parser) getBlockOnly(height int64) {
	p.NewBlockData = worker.GetBlockOnly(height)
}

func (p *Parser) getNetworkBlocksFromToHeight(heightFrom int64, heightTo int64) []types.BlockData {
	var networkBlocks []types.BlockData

	for i := heightFrom; i < heightTo; i++ {
		blockData := worker.GetBlockResult(i)

		networkBlocks = append(networkBlocks, types.BlockData{
			CosmosBlock: blockData.CosmosBlock,
			EvmBlock:    blockData.EvmBlock,
		})
	}

	return networkBlocks
}
