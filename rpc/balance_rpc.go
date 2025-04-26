package rpc

import (
	"encoding/json"
	"math/big"
	"strings"
)

func FetchBalance(address string) (*big.Int, error) {
	hexAddr, err := TronToHexAddress(address)
	if err != nil {
		return nil, err
	}

	params := []interface{}{hexAddr, "latest"}
	resp, err := post("eth_getBalance", params)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result string `json:"result"`
	}
	err = json.Unmarshal(resp, &rpcResp)
	if err != nil {
		return nil, err
	}

	// Balance is returned as hex string, example: "0x1234"
	balance := new(big.Int)
	balance.SetString(strings.TrimPrefix(rpcResp.Result, "0x"), 16)

	return balance, nil
}
