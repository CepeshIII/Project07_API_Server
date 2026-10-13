// Package credential provides cryptographic utilities for generating and hashing secure tokens.
package credential

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

// New generates a new opaque token and its SHA-256 hash.
// The plain token is returned to the client, while the hash can be safely stored in the database.
func New() (string, string) {
	plain := uuid.NewString()
	hash := Hash(plain)

	return plain, hash
}

// Hash returns the SHA-256 hash of the provided token as a hexadecimal string.
func Hash(value string) string {
	hash := sha256.Sum256([]byte(value))

	return hex.EncodeToString(hash[:])
}
