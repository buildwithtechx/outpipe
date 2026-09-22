package billing

import "outpipe.dev/outpipe/internal/security"

func VerifyHMACSHA256(payload []byte, signature, secret string) bool {
	return security.VerifyHMACSHA256(payload, signature, secret)
}
