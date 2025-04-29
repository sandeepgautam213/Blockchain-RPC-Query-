package models

import "time"

type TronTransaction struct {
	TxID         string
	Owner        string
	To           string
	Amount       int64
	ContractType string
	Timestamp    time.Time
}

type TronBlock struct {
	BlockNumber  int64
	Timestamp    time.Time
	Hash         string
	Transactions []TronTransaction
}
type TronscanTransaction struct {
	OwnerAddress string `json:"ownerAddress"`
	ToAddress    string `json:"toAddress"`
}

type TronscanResponse struct {
	Data []TronscanTransaction `json:"data"`
}

type EthTransaction struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EthBlock struct {
	Transactions []EthTransaction `json:"transactions"`
}
