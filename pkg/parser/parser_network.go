package parser

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/worker"
)

type NetworkData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func (p *Parser) getBlockFromNetwork(height int64, withTrx bool) {
	p.ChanelNewBlock = worker.GetBlockResult(height, withTrx)
}

func (p *Parser) getEvmBlock(height int64) evm.BlockEVM {
	return worker.GetEvmBlock(height)
}
