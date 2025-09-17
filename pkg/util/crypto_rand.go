package util

import (
	"crypto/rand"
	"math/big"
)

// CryptoRandInt generates a cryptographically secure random integer in [0, max)
func CryptoRandInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to a fixed value if crypto/rand fails
		return 0
	}
	return int(n.Int64())
}
