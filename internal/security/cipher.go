package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func EncryptAESGCM(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func DecryptAESGCM(ciphertext []byte, nonce []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid ciphertext nonce length")
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return plaintext, nil
}
