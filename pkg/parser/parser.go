package parser

import (
	"fmt"
	"sync"

	"bitbucket.org/decimalteam/api_fondation/types"
	"bitbucket.org/decimalteam/api_fondation/worker"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type BlockchainNetwork string

const (
	MainNet BlockchainNetwork = "https://node.decimalchain.com"
	TestNet BlockchainNetwork = "https://testnet-val.decimalchain.com"
	DevNet  BlockchainNetwork = "https://devnet-val.decimalchain.com"

	blockSubject = "block-subject"
)

type Parser struct {
	Interval         int // number in second for check new data
	Network          BlockchainNetwork
	IndexNode        string
	ParseServiceHost string
	NatsConfig       string
	Logger           *logrus.Logger
	NewBlockData     *types.BlockData
}

func NewParser(interval int, currNet BlockchainNetwork, indexNode, parseServiceHost, natsConfig string, logger *logrus.Logger) *Parser {

	// Setup config for decimal module
	config := sdk.GetConfig()
	worker.SetBech32Prefixes(config)
	worker.SetBip44CoinType(config)
	worker.RegisterBaseDenom()
	config.Seal()

	return &Parser{
		Interval:         interval,
		Network:          currNet,
		IndexNode:        indexNode,
		ParseServiceHost: parseServiceHost,
		NatsConfig:       natsConfig,
		Logger:           logger,
		NewBlockData:     nil,
	}
}

//TODO: 1. Na api Fondation parser.NewBlock nada zdelatiob esli block on ne ahodit v blokceine
//TODO: stob ne atvalilasi v panic a stob vernul nil

func (p *Parser) NewBlock(height int64) {

	p.getBlockFromIndexer(height)
	if p.NewBlockData != nil {
		return
	}

	p.getBlockFromNetwork(height)
}

func (p *Parser) GetNewBlockData(hFrom, hTo int64) []types.BlockData {
	return p.getBlocksFromToHeight(hFrom, hTo)
}

func (p *Parser) GetEvmBlock(height int64) {
	p.getEvmBlock(height)
}

func (p *Parser) GetBlockOnly(height int64) {
	p.getBlockOnly(height)
}

func getBlockFromNats(natsConfig string) (*types.BlockData, error) {
	var res *types.BlockData

	nc, err := nats.Connect(natsConfig)
	if err != nil {
		fmt.Printf("nats connect error: %v ", err)
		return res, err
	}
	nc.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	if _, err = nc.Subscribe(blockSubject, func(msg *nats.Msg) {
		wg.Done()
	}); err != nil {
		fmt.Printf("nats subscribe error: %v ", err)
		return res, err

	}

	wg.Wait()

	//TODO: get msgs from mats

	return res, nil
}
