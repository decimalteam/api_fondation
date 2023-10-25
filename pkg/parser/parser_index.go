package parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
	"encoding/json"
	"fmt"
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

	var dataBlock map[string]interface{}

	_ = json.Unmarshal(bytes, &dataBlock)

	fmt.Println(dataBlock)
}
