package parser

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/worker"
	"context"
)

type NetworkData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromNetwork(height int64) (*BlockData, error) {

	cosmosBlock := worker.GetBlockResult(height)

	evmBlock, err := evm.Parse(context.Background(), height)
	if err != nil {
		return nil, err
	}

	return &BlockData{
		CosmosBlock: cosmosBlock,
		EvmBlock:    evmBlock,
	}, nil
}