package utils

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcutil/base58"
)

func DecodeBase58Address(address string) (string, error) {
	if len(address) == 0 {
		return "", errors.New("address is empty")
	}
	addrBytes := base58.Decode(address)
	// addrBytes := base58.Decode(address)
	if len(addrBytes) == 0 {
		return "", errors.New("failed to decode address")
	}
	return hex.EncodeToString(addrBytes), nil
}
