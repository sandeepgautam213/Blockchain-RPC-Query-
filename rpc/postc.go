package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"crypto/sha256"

	"github.com/btcsuite/btcutil/base58"
	"github.com/joho/godotenv"
)

var rpcURL string

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error while loading .env file")
	}
	rpcURL = os.Getenv("RPC_URL")
	if rpcURL == "" {
		fmt.Println("rpc url not found")
	}
}

type TronTransaction struct {
	Owner string
	To    string
}

type TronBlock struct {
	Transactions []TronTransaction
}

func getLatestTronBlockNumber() (int64, error) {
	payload := map[string]interface{}{
		"method": "eth_blockNumber",
		"params": []interface{}{},
		"id":     1,
	}

	resp, err := postTronRPC(payload)
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

func getTronBlockByNumber(blockNum int64) (*TronBlock, error) {
	payload := map[string]interface{}{
		"method": "eth_getBlockByNumber",
		"params": []interface{}{fmt.Sprintf("0x%x", blockNum), true},
		"id":     1,
	}

	resp, err := postTronRPC(payload)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Result struct {
			Transactions []struct {
				From string `json:"from"`
				To   string `json:"to"`
			} `json:"transactions"`
		} `json:"result"`
	}
	err = json.Unmarshal(resp, &raw)
	if err != nil {
		return nil, err
	}

	block := &TronBlock{}
	for _, tx := range raw.Result.Transactions {
		block.Transactions = append(block.Transactions, TronTransaction{
			Owner: tx.From,
			To:    tx.To,
		})
	}

	return block, nil
}

func postTronRPC(payload map[string]interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// Tron address prefix for mainnet
var tronPrefix = []byte{0x41}

// Convert hex to base58 Tron address
func hexToTronAddress(hexAddr string) (string, error) {
	addrBytes, err := hex.DecodeString(strings.TrimPrefix(hexAddr, "0x"))
	if err != nil {
		return "", err
	}

	// Add Tron prefix 0x41
	fullAddr := append(tronPrefix, addrBytes...)

	// Add checksum
	firstHash := sha256.Sum256(fullAddr)
	secondHash := sha256.Sum256(firstHash[:])
	checksum := secondHash[:4]

	fullAddrWithChecksum := append(fullAddr, checksum...)

	return base58.Encode(fullAddrWithChecksum), nil
}
