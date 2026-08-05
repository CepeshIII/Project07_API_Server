package main

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

func NewInvitationToken() (string, string, error) {

	plainToken := uuid.New().String()

	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	return plainToken, hashToken, nil
}
