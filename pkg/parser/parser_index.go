package parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
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
	bytes := client.GetRequest(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	cl, err := client.New()
	if err != nil {
		fmt.Printf("create new client error: %v", err)
		return res, err
	}

	if err != nil {
		fmt.Printf("get hostname error: %v", err)
		return res, err
	}
	req.Header.Set("X-Worker", cl.Hostname)

	resp, err := cl.HttpClient.Do(req)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}
	defer func() {
		defErr := resp.Body.Close()
		if defErr != nil {
			fmt.Printf("http response close error: %v", err)
		}
	}()

	// Parse response
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}
	height, err := strconv.Atoi(string(bytes))
	if err != nil {
		fmt.Printf("get block from indexer error: %v", err)
		return res, err
	}

	res = worker.GetBlockResult(int64(height))

	return res, nil
}
