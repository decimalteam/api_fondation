package parser

import (
	"fmt"
	"strconv"

	"bitbucket.org/decimalteam/api_fondation/client"
)

type IndexData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(height int64) {
	//var res *types.BlockData

	url := fmt.Sprintf("%s/getBlock?height=%d", p.IndexNode, height)
	bytes := client.GetRequest(url)

	dataBlock, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
	}

	fmt.Println(dataBlock)
}
