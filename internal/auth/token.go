package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func NewToken(prefix string, size int) (string, error) {

	if size < 32 {
		return "", fmt.Errorf("token size must be at least 32 bytes")
	}

	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(value)

	if prefix == "" {
		return encoded, nil
	}

	return prefix + "_" + encoded, nil
}

func HashToken(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func EqualHash(expected, value string) bool {
	actual := HashToken(value)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
