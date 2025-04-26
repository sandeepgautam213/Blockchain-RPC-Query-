package handlers

import (
	"encoding/json"
	"net/http"
	"tron_rpc/rpc"
)

func CurrentBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	balanceBigInt, err := rpc.FetchBalance(address)
	if err != nil {
		http.Error(w, "Error fetching balance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	balanceInt := balanceBigInt.Int64()
	balanceTRX := float64(balanceInt) / 1e6

	resp := map[string]interface{}{
		"balance_trx": balanceTRX,
	}
	json.NewEncoder(w).Encode(resp)
}
