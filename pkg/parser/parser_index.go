package parser

import (
	"bitbucket.org/decimalteam/api_fondation/types"
	"fmt"
	"strconv"

	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/worker"
)

type IndexData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(indexerNode string) (*types.BlockData, error) {
	var res *types.BlockData

	url := fmt.Sprintf("%s/getWork", indexerNode)
	bytes := client.GetRequest(url)

	height, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	return worker.GetBlockResult(int64(height), true), nil
}
