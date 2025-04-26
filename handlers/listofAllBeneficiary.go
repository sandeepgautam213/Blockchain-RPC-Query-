package handlers

import (
	"encoding/json"
	"net/http"
	"tron_rpc/rpc"
)

func ListOfAllBeneficiary(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}
	beneficiaries, err := rpc.FetchBeneficiaries(address, 500)
	if err != nil {
		http.Error(w, "Error fetching beneficiaries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]interface{}{
		"address":       address,
		"beneficiaries": beneficiaries,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}
