package clients

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	http2 "github.com/tendermint/tendermint/rpc/client/http"
	"math/big"
	"net/http"
	"os"
)

func GetHostName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, nil
}

func GetHttpClient() *http.Client {
	return &http.Client{}
}

func GetWeb3Client(config *Config) (*ethclient.Client, error) {
	web3Client, err := ethclient.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	return web3Client, nil
}

func GetWeb3ChainId(web3Client *ethclient.Client) (*big.Int, error) {
	web3ChainId, err := web3Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return web3ChainId, nil
}

func GetRpcClient(config *Config, httpClient *http.Client) (*http2.HTTP, error) {
	rpcClient, err := http2.NewWithClient(config.RpcEndpoint, config.RpcEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func GetEthRpcClient(config *Config) (*rpc.Client, error) {
	ethRpcClient, err := rpc.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	return ethRpcClient, nil
}
