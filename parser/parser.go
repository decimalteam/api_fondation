package parser

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
