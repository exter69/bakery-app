package token

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate returns a cryptographically random hex-encoded token.
func Generate() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
