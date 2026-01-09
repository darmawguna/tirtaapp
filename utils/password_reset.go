package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func GenerateResetToken(bytesLen int) (string, error) {
	if bytesLen < 16 {
		return "", errors.New("bytesLen too small")
	}
	b := make([]byte, bytesLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// hex gampang untuk parsing di mobile + aman dibawa di URL
	return hex.EncodeToString(b), nil
}

func HashTokenSHA256Hex(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
