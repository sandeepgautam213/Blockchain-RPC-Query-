package handlers

import (
	"encoding/json"
	"net/http"
	"tron_rpc/rpc"
)

func PayerAddress(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Please mention address", http.StatusBadRequest)
		return
	}
	payers, err := rpc.FetchPayers(address, targetCount, maxDepth)
	if err != nil {
		http.Error(w, "Error fetching payers:"+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"address": address,
		"payers":  payers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
