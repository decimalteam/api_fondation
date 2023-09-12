package parser

import (
	"bitbucket.org/decimalteam/api_fondation/parser/cosmos"
	"encoding/json"
	"fmt"
	"github.com/tendermint/tendermint/abci/types"
	"io"
	"net/http"
	"time"
)

type BlockchainNetwork string

const (
	MainNet BlockchainNetwork = "https://node.decimalchain.com"
	TestNet BlockchainNetwork = "https://testnet-val.decimalchain.com"
	DevNet  BlockchainNetwork = "https://devnet-val.decimalchain.com"
)

type Parser struct {
	Interval         int // number in second for check new data
	Network          BlockchainNetwork
	IndexNode        string
	ParseServiceHost string
	NatsConfig       string
}

func NewParser(interval int, currNet BlockchainNetwork, indexNode, parseServiceHost, natsConfig string) *Parser {

	return &Parser{
		Interval:         interval,
		Network:          currNet,
		IndexNode:        indexNode,
		ParseServiceHost: parseServiceHost,
		NatsConfig:       natsConfig,
	}
}

func (p *Parser) NewBlock(ch chan cosmos.Block) {

	block, err := getBlockFromParser(address)
	if err != nil {
		return
	}

	ch <- block
}

func getBlockFromParser(address string) (cosmos.Block, error) {
	var res cosmos.Block

	blockDataResponse, err := downloadBlockData(address)
	if err != nil {
		fmt.Printf("block data request error: %v ", err)
		return res, err
	}

	res = cosmos.Block{
		height:          blockDataResponse.Block.Header.Height,
		date:            blockDataResponse.Block.Header.Time.Format(time.RFC3339Nano),
		hash:            blockDataResponse.BlockId.Hash,
		blockTime:       int(blockDataResponse.Block.Header.Time.Unix()),
		txsCount:        len(blockDataResponse.Block.Data.Txs),
		validatorsCount: len(blockDataResponse.Block.LastCommit.Signatures),
		proposerAddress: blockDataResponse.Block.Header.ProposerAddress,
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
