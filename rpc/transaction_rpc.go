package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
)

type Block struct {
	Number       string        `json:"number"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

func FetchPayers(address string, blockRange int64) ([]string, error) {
	payers := make(map[string]bool)
	err := scanBlocks(address, blockRange, func(from, to string) {
		if to == address {
			payers[from] = true
		}
	})
	return keys(payers), err
}

func FetchBeneficiaries(address string, blockRange int64) ([]string, error) {
	beneficiaries := make(map[string]bool)
	err := scanBlocks(address, blockRange, func(from, to string) {
		if from == address {
			beneficiaries[to] = true
		}
	})
	return keys(beneficiaries), err
}

func scanBlocks(address string, blockRange int64, handleTx func(from, to string)) error {
	latest, err := getLatestBlockNumber()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5) // limit concurrency to 5

	for i := latest; i >= latest-blockRange; i-- {
		wg.Add(1)
		sem <- struct{}{}

		go func(num int64) {
			defer func() {
				wg.Done()
				<-sem
			}()

			block, err := getBlockByNum(num)
			if err != nil {
				return
			}

			for _, tx := range block.Transactions {
				from := tx.From
				to := tx.To

				if from == "" || to == "" {
					continue
				}

				mu.Lock()
				handleTx(from, to)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func getLatestBlockNumber() (int64, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := postc(payload)
	if err != nil {
		return 0, err
	}

	var rpcResp struct {
		Result string `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	err = json.Unmarshal(resp, &rpcResp)
	if err != nil {
		return 0, err
	}

	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC Error: %s", rpcResp.Error.Message)
	}

	blockNum, ok := new(big.Int).SetString(rpcResp.Result[2:], 16)
	if !ok {
		return 0, fmt.Errorf("failed to parse block number")
	}

	return blockNum.Int64(), nil
}

func getBlockByNum(num int64) (*Block, error) {
	hexNum := fmt.Sprintf("0x%x", num) // convert block number to hex

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{hexNum, true},
		"id":      1,
	}

	resp, err := postc(payload)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result *Block `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	err = json.Unmarshal(resp, &rpcResp)
	if err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC Error: %s", rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
