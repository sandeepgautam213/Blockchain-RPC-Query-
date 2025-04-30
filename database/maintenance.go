package db

import (
	"fmt"
)

func EnforceMaxDBSize(maxSizeMB int64) error {
	var sizeMB float64
	err := DB.QueryRow(`
		SELECT pg_database_size(current_database()) / 1024 / 1024
	`).Scan(&sizeMB)
	if err != nil {
		return fmt.Errorf("failed to get DB size: %v", err)
	}

	if int64(sizeMB) > maxSizeMB {
		latestBlock, err := GetLatestStoredBlock()
		if err != nil {
			return err
		}
		return PruneOldData(latestBlock)
	}

	return nil
}
