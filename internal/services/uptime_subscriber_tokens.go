package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var subscriberEmailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func normalizeSubscriberEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || len(normalized) > 254 || !subscriberEmailPattern.MatchString(normalized) {
		return "", fmt.Errorf("valid email address is required")
	}
	return normalized, nil
}

func newSubscriberRawToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashSubscriberToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func signSubscriberPayload(key, payload string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func mintSubscriberToken(key, subscriberID, purpose string, expiresAt time.Time) (string, error) {
	raw, err := newSubscriberRawToken()
	if err != nil {
		return "", err
	}
	payload := strings.Join([]string{subscriberID, strconv.FormatInt(expiresAt.Unix(), 10), purpose, raw}, ".")
	return payload + "." + signSubscriberPayload(key, payload), nil
}

func mintUnsubscribeToken(key, subscriberID string, expiresAt time.Time) string {
	payload := strings.Join([]string{subscriberID, strconv.FormatInt(expiresAt.Unix(), 10), "unsubscribe"}, ".")
	return payload + "." + signSubscriberPayload(key, payload)
}

type verifiedSubscriberToken struct {
	subscriberID string
	expiresAt    time.Time
	purpose      string
}

func verifySubscriberToken(key, token, purpose string) (verifiedSubscriberToken, error) {
	var verified verifiedSubscriberToken
	parts := strings.Split(token, ".")
	if len(parts) < 4 {
		return verified, fmt.Errorf("invalid token")
	}
	payload := strings.Join(parts[:len(parts)-1], ".")
	payloadParts := strings.Split(payload, ".")
	if len(payloadParts) < 3 || payloadParts[2] != purpose {
		return verified, fmt.Errorf("invalid token")
	}
	if purpose == "confirm" && len(payloadParts) != 4 {
		return verified, fmt.Errorf("invalid token")
	}
	if purpose == "unsubscribe" && len(payloadParts) != 3 {
		return verified, fmt.Errorf("invalid token")
	}
	expected := signSubscriberPayload(key, payload)
	signature, err := hex.DecodeString(parts[len(parts)-1])
	if err != nil {
		return verified, fmt.Errorf("invalid token")
	}
	want, err := hex.DecodeString(expected)
	if err != nil {
		return verified, fmt.Errorf("invalid token")
	}
	if subtle.ConstantTimeCompare(signature, want) != 1 {
		return verified, fmt.Errorf("invalid token")
	}
	expiry, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return verified, fmt.Errorf("invalid token")
	}
	verified = verifiedSubscriberToken{subscriberID: payloadParts[0], expiresAt: time.Unix(expiry, 0).UTC(), purpose: payloadParts[2]}
	if verified.subscriberID == "" {
		return verifiedSubscriberToken{}, fmt.Errorf("invalid token")
	}
	return verified, nil
}

type subscriberCursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

func encodeSubscriberCursor(createdAt time.Time, id string) string {
	raw, err := json.Marshal(subscriberCursor{CreatedAt: createdAt.UTC(), ID: id})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeSubscriberCursor(cursor string) (time.Time, string, error) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return time.Time{}, "", nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	var decoded subscriberCursor
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	if decoded.ID == "" || decoded.CreatedAt.IsZero() {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	return decoded.CreatedAt, decoded.ID, nil
}

func hashRateLimitKey(key, kind, value string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(kind + ":" + value))
	return hex.EncodeToString(mac.Sum(nil))
}
