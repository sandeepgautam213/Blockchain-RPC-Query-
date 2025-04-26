package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"tron_rpc/handlers"

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

func main() {
	http.HandleFunc("/currentBalance", handlers.CurrentBalance)
	http.HandleFunc("/payerAddress", handlers.PayerAddress)
	http.HandleFunc("/listOfAllBeneficiary", handlers.ListOfAllBeneficiary)

	fmt.Println("Server started at : 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
