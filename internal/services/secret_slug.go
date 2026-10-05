package services

import (
	"fmt"
	"strings"
)

type SecretInputError string

func (e SecretInputError) Error() string { return string(e) }

func validateSecretSlug(raw string, maxLength int) error {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if len(slug) == 0 || len(slug) > maxLength {
		return SecretInputError(fmt.Sprintf("secret scope slug must contain 1 to %d characters", maxLength))
	}
	for _, character := range slug {
		if character != '-' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return SecretInputError("secret scope slug may contain only letters, digits, and hyphens")
		}
	}
	return nil
}
