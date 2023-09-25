package parser

import (
	"bitbucket.org/decimalteam/api_fondation/clients"
	"bitbucket.org/decimalteam/api_fondation/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/worker"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
	"io"
	"net/http"
	"strconv"
	"sync"
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

func (p *Parser) NewBlock(ch chan *BlockData) {
	var blockData = new(BlockData)

	indexNodeBlock, err := getBlockFromIndexer(p.IndexNode)
	if err != nil {
		return
	}
	blockData.IndexerBlock = indexNodeBlock

	parseServiceBlock, err := getBlockFromDataSource(p.ParseServiceHost)
	if err != nil {
		return
	}
	blockData.ParseServiceBlock = parseServiceBlock

	natsBlock, err := getBlockFromNats(p.NatsConfig)
	if err != nil {
		return
	}
	blockData.NatsBlock = natsBlock

	ch <- blockData
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

func getBlockFromDataSource(address string) (*cosmos.Block, error) {
	var res *cosmos.Block

	_, err := downloadBlockData(address)
	if err != nil {
		fmt.Printf("block data request error: %v ", err)
		return res, err
	}

	res = &cosmos.Block{
		//TODO: add mapping block data response to cosmos.Block
	}

	return res, nil
}

func downloadBlockData(path string) (*types.Response, error) {
	request, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get request error: %s", err)
	}

	request.Header = http.Header{
		"Content-type": {"application/json"},
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get request error: %s", err)
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error: %s", err)
	}

	var result types.Response
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
