package main

import (
	"fmt"
	"log"
	"net/http"
	db "tron_rpc/database"
	"tron_rpc/handlers"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
}

func main() {
	db.InitDB() // connect and create schema

	http.HandleFunc("/currentBalance", handlers.CurrentBalance)
	http.HandleFunc("/payerAddress", handlers.PayerAddress)
	http.HandleFunc("/listOfAllBeneficiary", handlers.ListOfAllBeneficiary)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
