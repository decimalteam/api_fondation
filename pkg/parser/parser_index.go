package parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/worker"
	"fmt"
	"strconv"
)

type IndexData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func getBlockFromIndexer(indexerNode string) (*cosmos.Block, error) {
	var res *cosmos.Block

	url := fmt.Sprintf("%s/getWork", indexerNode)
	bytes := client.GetRequest(url)

	height, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	res = worker.GetBlockResult(int64(height))

	return res, nil
}
