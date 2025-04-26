package rpc

import (
	"encoding/json"
	"sync"
)

type Block struct {
	Transactions []Transaction `json: "transactions"`
}
type Transaction struct {
	RawData struct {
		Contract []Contract `json : "contract`
	} `json : "raw_data"`
}
type Contract struct {
	Type      string `json :"type"`
	Parameter struct {
		Value struct {
			OwnerAddress string `json:"owner_address"`
			ToAddress    string `json: "to_address`
		} `json: "value"`
	} `json:"parameter"`
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
	sem := make(chan struct{}, 5)

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
				for _, c := range tx.RawData.Contract {
					if c.Type != "TransferContract" {
						continue
					}
					from := c.Parameter.Value.OwnerAddress
					to := c.Parameter.Value.ToAddress

					mu.Lock()
					handleTx(from, to)
					mu.Unlock()

				}
			}

		}(i)
	}
	wg.Wait()
	return nil
}

func getLatestBlockNumber() (int64, error) {
	resp, err := post("/wallet/getnowblock", nil)
	if err != nil {
		return 0, err
	}

	var result struct {
		BlockHeader struct {
			RawData struct {
				Number int64 `json: "number`
			} `json: "raw_data"`
		} `json: "block_header"`
	}
	err = json.Unmarshal(resp, &result)
	return result.BlockHeader.RawData.Number, err
}

func getBlockByNum(num int64) (*Block, error) {
	body := map[string]interface{}{"num": num}
	resp, err := post("/wallet/getblockbynum", body)
	if err != nil {
		return nil, err
	}
	var block Block
	err = json.Unmarshal(resp, &block)
	return &block, err

}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
