package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordMemory      = 64 * 1024
	passwordIterations  = 3
	passwordParallelism = 2
	passwordKeyLength   = 32
	passwordSaltLength  = 16
)

func HashPassword(password string) (string, error) {

	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("password is required")
	}

	salt := make([]byte, passwordSaltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	derived := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemory, passwordParallelism, passwordKeyLength)
	encode := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", passwordMemory, passwordIterations, passwordParallelism, encode(salt), encode(derived)), nil
}

func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")

	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return false
	}

	var memory, iterations, parallelism uint32

	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}

	if memory != passwordMemory || iterations != passwordIterations || parallelism != passwordParallelism {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])

	if err != nil || len(salt) == 0 {
		return false
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[4])

	if err != nil || len(expected) == 0 {
		return false
	}

	actual := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
