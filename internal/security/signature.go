package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignHMACSHA256(payload []byte, secret string) string {
	digest := hmac.New(sha256.New, []byte(secret))
	_, _ = digest.Write(payload)
	return hex.EncodeToString(digest.Sum(nil))
}

func VerifyHMACSHA256(payload []byte, signature, secret string) bool {
	expected, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	overall := hmac.New(sha256.New, []byte(secret))
	_, _ = overall.Write(payload)
	return hmac.Equal(overall.Sum(nil), expected)
}
