package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type TronscanTransaction struct {
	OwnerAddress string `json:"ownerAddress"`
	ToAddress    string `json:"toAddress"`
}

type TronscanResponse struct {
	Data []TronscanTransaction `json:"data"`
}

type EthTransaction struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EthBlock struct {
	Transactions []EthTransaction `json:"transactions"`
}

// Fetch payers: who sent to my address
func FetchPayers(address string, limit int) ([]string, error) {
	if strings.HasPrefix(address, "0x") {
		fmt.Println("Fatching Payers")
		return fetchEthPayers(address, limit)
	} else {
		fmt.Println("Fatching tron payers")
		return fetchTronPayers(address, limit)
	}
}

// Fetch beneficiaries: where my address sent to
func FetchBeneficiaries(address string, limit int) ([]string, error) {
	if strings.HasPrefix(address, "0x") {
		return fetchEthBeneficiaries(address, limit)
	} else {
		return fetchTronBeneficiaries(address, limit)
	}
}

func fetchTronPayers(address string, limit int) ([]string, error) {
	url := fmt.Sprintf("https://apilist.tronscanapi.com/api/transaction?address=%s&limit=%d&sort=-timestamp", address, limit)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var apiResp TronscanResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	payers := make(map[string]bool)
	for _, tx := range apiResp.Data {
		if tx.ToAddress == address {
			payers[tx.OwnerAddress] = true
		} else if tx.OwnerAddress == address {
			payers[tx.ToAddress] = true
		}

	}

	return keys(payers), nil
}

func fetchTronBeneficiaries(address string, limit int) ([]string, error) {
	url := fmt.Sprintf("https://apilist.tronscanapi.com/api/transaction?address=%s&limit=%d&sort=-timestamp", address, limit)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println("=== RAW Tronscan API Response ===")
	// // fmt.Println(string(body))
	// fmt.Println("===============================")

	var apiResp TronscanResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	beneficiaries := make(map[string]bool)
	for _, tx := range apiResp.Data {
		if tx.OwnerAddress == address {
			beneficiaries[tx.ToAddress] = true
		}
	}

	return keys(beneficiaries), nil
}

func fetchEthPayers(address string, limit int) ([]string, error) {
	latestBlock, err := getLatestEthBlockNumber()
	fmt.Println("Sandeep payers")
	fmt.Println(latestBlock)
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

func fetchEthBeneficiaries(address string, limit int) ([]string, error) {
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

func getEthBlockByNumber(num int64) (*EthBlock, error) {
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
		Result EthBlock `json:"result"`
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
