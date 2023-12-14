package client

import "os"

type Config struct {
	IndexerEndpoint string
	RpcEndpoint     string
	Web3Endpoint    string
	WorkersCount    int
}

func GetConfig() *Config {
	return &Config{
		IndexerEndpoint: os.Getenv("INDEXER_URL"),
		RpcEndpoint:     os.Getenv("RPC_URL"),
		Web3Endpoint:    os.Getenv("WEB3_URL"),
		WorkersCount:    1,
	}
}
