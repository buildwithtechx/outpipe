package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RelayClaims struct {
	Sub            string `json:"sub"`
	Org            string `json:"org"`
	MaxTunnels     int    `json:"max_tunnels,omitempty"`
	MaxConnections int    `json:"max_conns,omitempty"`
	BandwidthBytes int64  `json:"bandwidth_bytes,omitempty"`
	Exp            int64  `json:"exp"`
	Iat            int64  `json:"iat"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func SignRelayToken(claims RelayClaims, secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("secret is required for signing")
	}

	if claims.Sub == "" || claims.Org == "" {
		return "", fmt.Errorf("sub (agent_id) and org are required")
	}

	if claims.Iat == 0 {
		claims.Iat = time.Now().Unix()
	}

	if claims.Exp == 0 {
		claims.Exp = time.Now().Add(15 * time.Minute).Unix()
	}

	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("encode jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode jwt claims: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	sig := hmacSHA256([]byte(signingInput), []byte(secret))
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigB64, nil
}

func VerifyRelayToken(tokenString, secret string) (RelayClaims, error) {
	if strings.TrimSpace(secret) == "" {
		return RelayClaims{}, fmt.Errorf("secret is required for verification")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return RelayClaims{}, fmt.Errorf("invalid jwt token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := hmacSHA256([]byte(signingInput), []byte(secret))

	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return RelayClaims{}, fmt.Errorf("decode jwt signature: %w", err)
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return RelayClaims{}, fmt.Errorf("invalid jwt signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return RelayClaims{}, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims RelayClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return RelayClaims{}, fmt.Errorf("unmarshal jwt claims: %w", err)
	}

	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return RelayClaims{}, fmt.Errorf("jwt token is expired")
	}

	if claims.Sub == "" || claims.Org == "" {
		return RelayClaims{}, fmt.Errorf("invalid jwt claims: missing sub or org")
	}

	return claims, nil
}

func hmacSHA256(data, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return h.Sum(nil)
}
