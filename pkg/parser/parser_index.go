package parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"encoding/json"
	"fmt"
)

type IndexData struct {
	Height  string        `json:"height"`
	Data    *cosmos.Block `json:"data"`
	EvmData *evm.BlockEVM `json:"evmData"`
}

func (p *Parser) getBlockFromIndexer(height int64) {
	//var res *types.BlockData

	url := fmt.Sprintf("%s/getBlock?height=%d", p.IndexNode, height)
	bytes := client.GetRequest(url)

	var dataBlock IndexData

	_ = json.Unmarshal(bytes, &dataBlock)

	fmt.Println(dataBlock)
}
