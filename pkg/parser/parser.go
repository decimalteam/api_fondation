package parser

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/evm"
	"bitbucket.org/decimalteam/api_fondation/types"
	"bitbucket.org/decimalteam/api_fondation/worker"
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
}

type BlockData struct {
	CosmosBlock       *cosmos.Block
	EvmBlock          *types.BlockEVM
	IndexerBlock      *cosmos.Block
	ParseServiceBlock *cosmos.Block
	NatsBlock         *cosmos.Block
}

func NewParser(interval int, currNet BlockchainNetwork, indexNode, parseServiceHost, natsConfig string, logger *logrus.Logger) *Parser {

	return &Parser{
		Interval:         interval,
		Network:          currNet,
		IndexNode:        indexNode,
		ParseServiceHost: parseServiceHost,
		NatsConfig:       natsConfig,
		Logger:           logger,
	}
}

func (p *Parser) NewBlock(ch chan *cosmos.Block) {
	indexNodeBlock, err := getBlockFromIndexer(p.IndexNode)
	if err != nil {
		return
	}
	ch <- indexNodeBlock

	parseServiceBlock, err := getBlockFromDataSource(p.ParseServiceHost)
	if err != nil {
		return
	}
	ch <- parseServiceBlock

	natsBlock, err := getBlockFromNats(p.NatsConfig)
	if err != nil {
		return
	}
	ch <- natsBlock
}

func getBlockFromNats(natsConfig string) (*cosmos.Block, error) {
	var res *cosmos.Block

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

func getBlockFromDataSource(address string) (*BlockData, error) {
	var res *cosmos.Block

	bytes := client.GetRequest(address)

	height, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	cosmosBlock := worker.GetBlockResult(int64(height))

	evmBlock, err := evm.Parse(context.Background(), int64(height))
	if err != nil {
		return nil, err
	}

	return &BlockData{
		CosmosBlock: cosmosBlock,
		EvmBlock:    evmBlock,
	}, nil
}
