package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type SecretBox struct{ aead cipher.AEAD }

func NewSecretBox(key []byte) (*SecretBox, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, fmt.Errorf("create secret cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)

	if err != nil {
		return nil, fmt.Errorf("create secret box: %w", err)
	}

	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) Seal(value string) (string, error) {
	nonce := make([]byte, b.aead.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate secret nonce: %w", err)
	}

	ciphertext := b.aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (b *SecretBox) Open(value string) (string, error) {
	ciphertext, err := base64.RawStdEncoding.DecodeString(value)

	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	nonceSize := b.aead.NonceSize()

	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("encrypted secret is too short")
	}

	plaintext, err := b.aead.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)

	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}

	return string(plaintext), nil
}
