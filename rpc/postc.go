package rpc

import (
	"fmt"
	"os"

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
