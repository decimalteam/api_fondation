package client

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	http2 "github.com/tendermint/tendermint/rpc/client/http"
	"math/big"
	"net/http"
	"os"
)

type Client struct {
	HttpClient   *http.Client
	Web3Client   *ethclient.Client
	RpcClient    *http2.HTTP
	EthRpcClient *rpc.Client
	Web3ChainId  *big.Int
	Hostname     string
}

func GetHttpClient() *http.Client {
	return &http.Client{}
}

func New() (*Client, error) {
	config := GetConfig()

	httpClient := &http.Client{}

	web3Client, err := ethclient.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	rpcClient, err := http2.NewWithClient(config.RpcEndpoint, config.RpcEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	ethRpcClient, err := rpc.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	web3ChainId, err := web3Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	var hostname string
	hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	return &Client{
		HttpClient:   httpClient,
		Web3Client:   web3Client,
		RpcClient:    rpcClient,
		EthRpcClient: ethRpcClient,
		Web3ChainId:  web3ChainId,
		Hostname:     hostname,
	}, nil
}
