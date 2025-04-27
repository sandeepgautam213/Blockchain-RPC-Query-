package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"tron_rpc/rpc"
)

func ListOfAllBeneficiary(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address query parameter is required", http.StatusBadRequest)
		return
	}

	// Fetch payers
	payers, err := rpc.FetchPayers(address, 50)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching payers: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch beneficiaries
	beneficiaries, err := rpc.FetchBeneficiaries(address, 50)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching beneficiaries: %v", err), http.StatusInternalServerError)
		return
	}

	//Fetch balance
	balanceBigInt, err := rpc.FetchBalance(address)
	balanceFormatted := "N/A"
	if err == nil {
		balanceInt := balanceBigInt.Int64()
		balanceTRX := float64(balanceInt) / 1e6 // For TRON (TRX)
		balanceFormatted = fmt.Sprintf("%.6f TRX", balanceTRX)
	}

	// Build the formatted response
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Input Wallet Address: %s\n", address))
	output.WriteString(fmt.Sprintf("Current Balance: %s\n", balanceFormatted))
	output.WriteString("\nPayers:\n")
	for _, payer := range payers {
		output.WriteString(fmt.Sprintf("- %s\n", payer))
	}
	output.WriteString("\nBeneficiaries:\n")
	for _, beneficiary := range beneficiaries {
		output.WriteString(fmt.Sprintf("- %s\n", beneficiary))
	}

	// Set header and write output
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(output.String()))
}
