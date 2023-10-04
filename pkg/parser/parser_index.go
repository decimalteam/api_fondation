package parser

import (
	"bitbucket.org/decimalteam/api_fondation/clients"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"bitbucket.org/decimalteam/api_fondation/worker"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type IndexData struct {
	Height  string `json:"height"`
	Data    string `json:"data"`
	EvmData string `json:"evmData"`
}

func getBlockFromIndexer(indexerNode string) (*cosmos.Block, error) {
	var res *cosmos.Block

	url := fmt.Sprintf("%s/getWork", indexerNode)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	hostname, err := clients.GetHostName()
	if err != nil {
		fmt.Printf("get hostname error: %v", err)
		return res, err
	}
	req.Header.Set("X-Worker", hostname)

	clients.GetHttpClient()
	resp, err := clients.GetHttpClient().Do(req)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}
	defer resp.Body.Close()

	// Parse response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}
	height, err := strconv.Atoi(string(bodyBytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	res = worker.GetBlockResult(int64(height))

	return res, nil
}
