package db

import (
	"time"
)

func SaveBlockAndTx(
	blockNumber int64,
	blockTime time.Time,
	blockHash string,
	txID string,
	from string,
	to string,
	amount int64,
	contractType string,
) error {
	_, err := DB.Exec(`
		INSERT INTO tron_blocks (block_number, timestamp, block_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`,
		blockNumber, blockTime, blockHash,
	)

	if err != nil {
		return err
	}

	_, err = DB.Exec(`
		INSERT INTO tron_transactions (tx_id, block_number, from_address, to_address, amount, contract_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		txID, blockNumber, from, to, amount, contractType, blockTime,
	)

	return err
}
