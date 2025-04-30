package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	db "tron_rpc/database"
	"tron_rpc/models"
)

// Fetch payers: who sent to my address
func FetchPayers(address string, targetCount int, maxDepth int) ([]string, error) {
	if strings.HasPrefix(address, "0x") {
		fmt.Println("Fatching Payers")
		return fetchEthPayers(address, targetCount, maxDepth)
	} else {
		fmt.Println("Fatching tron payers")
		return fetchTronPayers(address, targetCount, maxDepth)
	}
}

// Fetch beneficiaries: where my address sent to
func FetchBeneficiaries(address string, targetCount int, maxDepth int) ([]string, error) {
	if strings.HasPrefix(address, "0x") {
		return fetchEthBeneficiaries(address, targetCount, maxDepth)
	} else {
		return fetchTronBeneficiaries(address, targetCount, maxDepth)
	}
}
func fetchTronPayers(address string, targetCount, maxDepth int) ([]string, error) {
	rows, err := db.DB.Query(`
		SELECT from_address FROM tron_transactions
		WHERE to_address = $1
		ORDER BY block_number DESC
		LIMIT $2
	`, address, maxDepth*20) // scale factor
	if err != nil {
		return nil, fmt.Errorf("payer query failed: %v", err)
	}
	defer rows.Close()

	payers := make(map[string]struct{})
	for rows.Next() {
		var from string
		if err := rows.Scan(&from); err != nil {
			continue
		}
		payers[from] = struct{}{}
		if len(payers) >= targetCount {
			break
		}
	}

	return keysFromMap(payers), nil
}

func fetchTronBeneficiaries(address string, targetCount, maxDepth int) ([]string, error) {
	rows, err := db.DB.Query(`
		SELECT to_address FROM tron_transactions
		WHERE from_address = $1
		ORDER BY block_number DESC
		LIMIT $2
	`, address, maxDepth*20)
	if err != nil {
		return nil, fmt.Errorf("beneficiary query failed: %v", err)
	}
	defer rows.Close()

	beneficiaries := make(map[string]struct{})
	for rows.Next() {
		var to string
		if err := rows.Scan(&to); err != nil {
			continue
		}
		beneficiaries[to] = struct{}{}
		if len(beneficiaries) >= targetCount {
			break
		}
	}

	return keysFromMap(beneficiaries), nil
}
func keysFromMap(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func fetchEthPayers(address string, limit int, maxDepth int) ([]string, error) {
	latestBlock, err := getLatestEthBlockNumber()
	// fmt.Println("Sandeep payers")
	// fmt.Println(latestBlock)
	if err != nil {
		return nil, err
	}

	payers := make(map[string]bool)

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5)

	for i := 0; i < limit; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer func() {
				wg.Done()
				<-sem
			}()

			blockNumber := latestBlock - int64(i)
			block, err := getEthBlockByNumber(blockNumber)
			if err != nil {
				return
			}

			for _, tx := range block.Transactions {
				if strings.EqualFold(tx.To, address) {
					mu.Lock()
					payers[tx.From] = true
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()

	return keys(payers), nil
}

func fetchEthBeneficiaries(address string, limit int, maxDepth int) ([]string, error) {
	latestBlock, err := getLatestEthBlockNumber()
	if err != nil {
		return nil, err
	}

	beneficiaries := make(map[string]bool)

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5)

	for i := 0; i < limit; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer func() {
				wg.Done()
				<-sem
			}()

			blockNumber := latestBlock - int64(i)
			block, err := getEthBlockByNumber(blockNumber)
			if err != nil {
				return
			}

			for _, tx := range block.Transactions {
				if strings.EqualFold(tx.From, address) {
					mu.Lock()
					beneficiaries[tx.To] = true
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()

	return keys(beneficiaries), nil
}

func getLatestEthBlockNumber() (int64, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}

	resp, err := postEthRPC(payload)
	if err != nil {
		return 0, err
	}

	var rpcResp struct {
		Result string `json:"result"`
	}
	err = json.Unmarshal(resp, &rpcResp)
	if err != nil {
		return 0, err
	}

	var blockNumber int64
	fmt.Sscanf(rpcResp.Result, "0x%x", &blockNumber)
	return blockNumber, nil
}

func getEthBlockByNumber(num int64) (*models.EthBlock, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{fmt.Sprintf("0x%x", num), true},
		"id":      1,
	}

	resp, err := postEthRPC(payload)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result models.EthBlock `json:"result"`
	}
	err = json.Unmarshal(resp, &rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp.Result, nil
}

func postEthRPC(payload map[string]interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

const (
	batchSize     = 10
	maxConcurrent = 5
)

func StartBlockSyncer() {
	sem := make(chan struct{}, maxConcurrent) // limit concurrency

	for {
		latestChainBlock, err := getLatestTronBlockNumber()
		if err != nil {
			log.Printf("Failed to get latest block: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		latestStored, err := db.GetLatestStoredBlock()
		if err != nil {
			log.Printf("Failed to get latest stored block: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if latestStored >= latestChainBlock {
			time.Sleep(5 * time.Second)
			continue
		}

		var wg sync.WaitGroup

		for b := latestStored + 1; b <= latestChainBlock; b += batchSize {
			end := b + batchSize - 1
			if end > latestChainBlock {
				end = latestChainBlock
			}

			wg.Add(1)
			sem <- struct{}{} // acquire

			go func(start, end int64) {
				defer wg.Done()
				defer func() { <-sem }() // release

				for i := start; i <= end; i++ {
					block, err := getTronBlockByNumber(int64(i))
					if err != nil {
						log.Printf("Error fetching block %d: %v", i, err)
						return
					}
					if err := db.StoreBlockAndTx(*block); err != nil {
						log.Printf("Error storing block %d: %v", i, err)
						return
					}
					if err := db.EnforceMaxDBSize(6000); err != nil {
						log.Printf("DB size enforcement failed: %v", err)
					}
					log.Printf("Synced block %d", i)
				}
			}(b, end)
		}

		wg.Wait()
	}
}
