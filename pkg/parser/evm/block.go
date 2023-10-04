package evm

import (
	"bitbucket.org/decimalteam/api_fondation/types"
	"context"
	"math/big"

	"bitbucket.org/decimalteam/api_fondation/clients"
	"bitbucket.org/decimalteam/api_fondation/worker"
	web3hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	web3types "github.com/ethereum/go-ethereum/core/types"
)

func Parse(ctx context.Context, height int64) (*types.BlockEVM, error) {
	web3BlockChan := make(chan *web3types.Block)

	web3Client, err := clients.GetWeb3Client(clients.GetConfig())
	if err != nil {
		return nil, err
	}

	go worker.FetchBlockWeb3(ctx, web3Client, height, web3BlockChan)

	web3Block := <-web3BlockChan
	web3Body := web3Block.Body()

	var web3ChainId *big.Int
	web3ChainId, err = clients.GetWeb3ChainId(web3Client)
	if err != nil {
		return nil, err
	}

	var web3Transactions []*types.TransactionEVM
	web3Transactions, err = getWeb3Transactions(web3Body, web3ChainId, web3Block)
	if err != nil {
		return nil, err
	}

	ethRpcClient, err := clients.GetEthRpcClient(clients.GetConfig())
	if err != nil {
		return nil, err
	}

	web3ReceiptsChan := make(chan web3types.Receipts)
	go worker.FetchBlockTxReceiptsWeb3(ethRpcClient, web3Block, web3ReceiptsChan)
	web3Receipts := <-web3ReceiptsChan

	return &types.BlockEVM{
		Header:       web3Block.Header(),
		Transactions: web3Transactions,
		Uncles:       web3Body.Uncles,
		Receipts:     web3Receipts,
	}, nil
}

func getWeb3Transactions(web3Body *web3types.Body, web3ChainId *big.Int, web3Block *web3types.Block) ([]*types.TransactionEVM, error) {
	web3Transactions := make([]*types.TransactionEVM, len(web3Body.Transactions))

	for i, tx := range web3Body.Transactions {
		msg, err := tx.AsMessage(web3types.NewLondonSigner(web3ChainId), nil)
		if err != nil {
			return nil, err
		}

		web3Transactions[i] = &types.TransactionEVM{
			Type:             web3hexutil.Uint64(tx.Type()),
			Hash:             tx.Hash(),
			Nonce:            web3hexutil.Uint64(tx.Nonce()),
			BlockHash:        web3Block.Hash(),
			BlockNumber:      web3hexutil.Uint64(web3Block.NumberU64()),
			TransactionIndex: web3hexutil.Uint64(uint64(i)),
			From:             msg.From(),
			To:               msg.To(),
			Value:            (*web3hexutil.Big)(msg.Value()),
			Data:             msg.Data(),
			Gas:              web3hexutil.Uint64(msg.Gas()),
			GasPrice:         (*web3hexutil.Big)(msg.GasPrice()),
			ChainId:          (*web3hexutil.Big)(tx.ChainId()),
			AccessList:       msg.AccessList(),
			GasTipCap:        (*web3hexutil.Big)(msg.GasTipCap()),
			GasFeeCap:        (*web3hexutil.Big)(msg.GasFeeCap()),
		}
	}

	return web3Transactions, nil
}
