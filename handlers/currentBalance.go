package handlers

import (
	"fmt"
	"net/http"
	"tron_rpc/rpc"
)

func CurrentBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address is required", http.StatusBadRequest)
		return
	}

	balance, err := rpc.FetchBalance(address)
	if err != nil {
		http.Error(w, "Error fetching balance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"balance": "%s"}`, balance.String())))
}
