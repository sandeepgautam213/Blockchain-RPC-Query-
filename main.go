package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	db "tron_rpc/database"
	"tron_rpc/handlers"
	"tron_rpc/rpc"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not loaded")
	}
}

func main() {
	db.InitDB() // Connect and create schema

	// Start background block syncer
	go rpc.StartBlockSyncer()
	log.Println("Started block syncer")

	// Schedule periodic DB cleanup
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := db.EnforceMaxDBSize(500); err != nil {
				log.Printf("Scheduled DB cleanup error: %v", err)
			} else {
				log.Println("Scheduled DB cleanup completed")
			}
		}
	}()
	log.Println("Scheduled DB cleanup initialized")

	// API routes
	http.HandleFunc("/currentBalance", handlers.CurrentBalance)
	http.HandleFunc("/payerAddress", handlers.PayerAddress)
	http.HandleFunc("/listOfAllBeneficiary", handlers.ListOfAllBeneficiary)

	log.Println("Server started at :8080")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	//Graceful shutdown handler
	waitForShutdown()
}

// Capture CTRL+C to allow cleanup
func waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")
	// Close DB or cleanup here if needed
}
