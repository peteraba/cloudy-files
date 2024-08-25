package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// RandomHex generates an n-length, random hex string.
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n/2) //nolint:mnd // (n/2 because each byte represents 2 hex digits)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generating random hex: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
