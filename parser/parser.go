package parser

import "bitbucket.org/decimalteam/api_fondation/parser/cosmos"

const (
	mainnet = "https://node.decimalchain.com"
	testnet = "https://testnet-val.decimalchain.com"
	devnet  = "https://devnet-val.decimalchain.com"
)

type Parser struct {
	Interval         int    // number in second for check new data
	Network          string // Name of network work mainnet, testnet
	IndexNode        string
	ParseServiceHost string
	NatsConfig       string
}

func NewParser(indexNode, parseServiceHost, natsConfig *string) *Parser {

	return &Parser{
		Interval:         0,
		Network:          "",
		IndexNode:        "",
		ParseServiceHost: "",
		NatsConfig:       "",
	}
}

func (p *Parser) newBlock() *cosmos.Block {

	return &cosmos.Block{}
}
