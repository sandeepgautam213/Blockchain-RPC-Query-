package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"tron_rpc/models"

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

func getTronBlockByNumber(blockNum int64) (*models.TronBlock, error) {
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
			Number       string `json:"number"`
			Timestamp    string `json:"timestamp"`
			Hash         string `json:"hash"`
			Transactions []struct {
				Hash           string `json:"hash"`
				From           string `json:"from"`
				To             string `json:"to"`
				Value          string `json:"value"`
				Input          string `json:"input"`
				BlockTimestamp string `json:"timestamp"` // fallback if present
			} `json:"transactions"`
		} `json:"result"`
	}

	err = json.Unmarshal(resp, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block response: %w", err)
	}

	blockNumber, err := strconv.ParseInt(raw.Result.Number, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block number: %w", err)
	}

	timestampUnix, err := strconv.ParseInt(raw.Result.Timestamp, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block timestamp: %w", err)
	}
	blockTimestamp := time.UnixMilli(timestampUnix)

	block := &models.TronBlock{
		BlockNumber:  blockNumber,
		Timestamp:    blockTimestamp,
		Hash:         raw.Result.Hash,
		Transactions: []models.TronTransaction{},
	}

	for _, tx := range raw.Result.Transactions {
		amount, _ := strconv.ParseInt(tx.Value, 0, 64)

		txTime := blockTimestamp
		if tx.BlockTimestamp != "" {
			if t, err := strconv.ParseInt(tx.BlockTimestamp, 0, 64); err == nil {
				txTime = time.UnixMilli(t)
			}
		}

		block.Transactions = append(block.Transactions, models.TronTransaction{
			TxID:         tx.Hash,
			Owner:        tx.From,
			To:           tx.To,
			Amount:       amount,
			ContractType: detectContractType(tx.Input),
			Timestamp:    txTime,
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
func detectContractType(input string) string {
	if len(input) < 10 {
		return "transfer"
	}
	// Example: Detect based on known function selectors
	switch input[:10] {
	case "0xa9059cbb":
		return "TRC20 transfer"
	default:
		return "contract call"
	}
}
