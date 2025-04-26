package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

func TronToHexAddress(tronAddr string) (string, error) {
	if tronAddr == "" {
		return "", errors.New("address is empty")
	}
	if !strings.HasPrefix(tronAddr, "T") {
		return "", errors.New("invalid TRON address prefix")
	}

	decoded := base58.Decode(tronAddr)
	if len(decoded) != 25 {
		return "", errors.New("invalid address length after base58 decoding")
	}

	// Verify checksum
	body := decoded[:21]
	checksum := decoded[21:]

	hash0 := sha256.Sum256(body)
	hash1 := sha256.Sum256(hash0[:])

	if !equal(checksum, hash1[:4]) {
		return "", errors.New("invalid checksum")
	}

	// Now, body[0] == 0x41 (mainnet prefix)
	if body[0] != 0x41 {
		return "", errors.New("invalid TRON address prefix byte")
	}

	addressHex := hex.EncodeToString(body)
	return "0x" + addressHex, nil
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
