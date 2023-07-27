package clients

import (
	"api_fondation"
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	http2 "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/valyala/fasthttp"
	"math/big"
	"net/http"
	"os"
)

func getHostName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, nil
}

func GetHttpClient() *fasthttp.Client {
	return &fasthttp.Client{}
}

func GetWeb3Client(config *api_fondation.Config) (*ethclient.Client, error) {
	web3Client, err := ethclient.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	return web3Client, nil
}

func getWeb3ChainId(web3Client *ethclient.Client) (*big.Int, error) {
	web3ChainId, err := web3Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return web3ChainId, nil
}

func GetRpcClient(config *api_fondation.Config, httpClient *http.Client) (*http2.HTTP, error) {
	rpcClient, err := http2.NewWithClient(config.RpcEndpoint, config.RpcEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func GetEthRpcClient(config *api_fondation.Config) (*rpc.Client, error) {
	ethRpcClient, err := rpc.Dial(config.Web3Endpoint)
	if err != nil {
		return nil, err
	}

	return ethRpcClient, nil
}
