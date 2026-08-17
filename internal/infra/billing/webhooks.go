package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func VerifyHMACSHA256(payload []byte, signature, secret string) bool {
	digest := hmac.New(sha256.New, []byte(secret))
	_, _ = digest.Write(payload)
	expected, err := hex.DecodeString(signature)

	if err != nil {
		return false
	}

	return hmac.Equal(digest.Sum(nil), expected)
}
