package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

func postc(payload map[string]interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("bad status: " + resp.Status + ", body: " + string(bodyBytes))
	}

	return ioutil.ReadAll(resp.Body)
}
