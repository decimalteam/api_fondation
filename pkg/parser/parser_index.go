package parser

import (
	"context"
	"fmt"
	"strconv"

	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/worker"
)

type IndexData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(indexerNode string) (*BlockData, error) {
	var res *BlockData

	url := fmt.Sprintf("%s/getWork", indexerNode)
	bytes := client.GetRequest(url)

	height, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	cosmosBlock := worker.GetBlockResult(int64(height))

	evmBlock, err := evm.Parse(context.Background(), int64(height))
	if err != nil {
		return nil, err
	}

	return &BlockData{
		CosmosBlock: cosmosBlock,
		EvmBlock:    evmBlock,
	}, nil
}
