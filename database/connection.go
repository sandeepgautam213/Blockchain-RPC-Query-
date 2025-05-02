package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"tron_rpc/models"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const startBlockNum = 71779551

func InitDB() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("DB not reachable: %v", err)
	}

	if err = initSchema(DB); err != nil {
		log.Fatalf("Schema init failed: %v", err)
	}

	log.Println("Database connected and schema ensured.")
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tron_blocks (
		block_number BIGINT PRIMARY KEY,
		timestamp TIMESTAMP,
		block_hash TEXT
	);

	CREATE TABLE IF NOT EXISTS tron_transactions (
		tx_id TEXT PRIMARY KEY,
		block_number BIGINT REFERENCES tron_blocks(block_number),
		from_address TEXT,
		to_address TEXT,
		amount BIGINT,
		contract_type TEXT,
		timestamp TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_to_address ON tron_transactions(to_address);
	CREATE INDEX IF NOT EXISTS idx_from_address ON tron_transactions(from_address);
	`
	_, err := db.Exec(schema)
	return err
}
func BlockExists(blockNum int64) bool {
	var exists bool
	err := DB.QueryRow(`SELECT EXISTS (SELECT 1 FROM tron_blocks WHERE block_number = $1)`, blockNum).Scan(&exists)
	if err != nil {
		log.Printf("BlockExists error: %v", err)
		return false
	}
	return exists
}

func StoreBlockAndTx(block models.TronBlock) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx error: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO tron_blocks (block_number, timestamp, block_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, block.BlockNumber, block.Timestamp, block.Hash)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert block error: %v", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO tron_transactions (tx_id, block_number, from_address, to_address, amount, contract_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare tx stmt error: %v", err)
	}
	defer stmt.Close()

	for _, t := range block.Transactions {
		_, err := stmt.Exec(t.TxID, block.BlockNumber, t.Owner, t.To, t.Amount, t.ContractType, t.Timestamp)
		if err != nil {
			log.Printf("failed to insert tx %s: %v", t.TxID, err)
			// continue on individual tx error
		}
	}

	return tx.Commit()
}

func PruneOldData(currentBlock int64) error {
	cutoff := currentBlock - 500000
	if cutoff <= 0 {
		return nil
	}
	_, err := DB.Exec(`
		DELETE FROM tron_transactions WHERE block_number < $1;
		DELETE FROM tron_blocks WHERE block_number < $1;
	`, cutoff, cutoff)
	return err
}
func GetLatestStoredBlock() (int64, error) {
	var blockNum sql.NullInt64
	err := DB.QueryRow(`SELECT MAX(block_number) FROM tron_blocks`).Scan(&blockNum)
	if err != nil {
		return 0, err
	}
	if blockNum.Valid {
		return blockNum.Int64, nil
	}
	// Table is empty; return default block number (e.g., 0)
	return startBlockNum, nil
}
