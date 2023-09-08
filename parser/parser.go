package parser

import "bitbucket.org/decimalteam/api_fondation/parser/cosmos"

type BlockchainNetwork string

const (
	MainNet BlockchainNetwork = "https://node.decimalchain.com"
	TestNet BlockchainNetwork = "https://testnet-val.decimalchain.com"
	DevNet  BlockchainNetwork = "https://devnet-val.decimalchain.com"
)

type Parser struct {
	Interval         int               // number in second for check new data
	Network          BlockchainNetwork // Name of network work mainnet, testnet
	IndexNode        string
	ParseServiceHost string
	NatsConfig       string
}

func NewParser(interval int, indexNode, parseServiceHost, natsConfig *string) *Parser {

	return &Parser{
		Interval:         interval,
		Network:          "",
		IndexNode:        "",
		ParseServiceHost: "",
		NatsConfig:       "",
	}
}

func (p *Parser) newBlock() *cosmos.Block {

	return &cosmos.Block{}
}
