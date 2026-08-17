package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type TunnelTokenService struct {
	tokens repositories.TunnelTokenRepository
	now    func() time.Time
}

type TunnelTokenCredential struct {
	Token  models.TunnelToken
	Scopes []string
}

func NewTunnelTokenService(tokens repositories.TunnelTokenRepository) (*TunnelTokenService, error) {

	if tokens == nil {
		return nil, fmt.Errorf("tunnel token repository is required")
	}

	return &TunnelTokenService{tokens: tokens, now: time.Now}, nil
}

func (s *TunnelTokenService) Create(ctx context.Context, tunnelID, name string, scopes []string, expiresAt *time.Time) (string, models.TunnelToken, error) {

	if tunnelID == "" || strings.TrimSpace(name) == "" {
		return "", models.TunnelToken{}, fmt.Errorf("tunnel id and token name are required")
	}

	prefix, err := auth.NewToken("cdt", 8)

	if err != nil {
		return "", models.TunnelToken{}, err
	}

	secret, err := auth.NewToken("", 32)

	if err != nil {
		return "", models.TunnelToken{}, err
	}

	scopeJSON, err := json.Marshal(scopes)

	if err != nil {
		return "", models.TunnelToken{}, fmt.Errorf("encode tunnel token scopes: %w", err)
	}

	token := models.TunnelToken{TunnelID: tunnelID, Name: strings.TrimSpace(name), Prefix: prefix, TokenHash: auth.HashToken(secret), Scopes: string(scopeJSON), ExpiresAt: expiresAt}

	if err := s.tokens.Create(ctx, &token); err != nil {
		return "", models.TunnelToken{}, fmt.Errorf("create tunnel token: %w", err)
	}

	return prefix + "." + secret, token, nil
}

func (s *TunnelTokenService) Authenticate(ctx context.Context, raw string) (TunnelTokenCredential, error) {
	parts := strings.SplitN(raw, ".", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return TunnelTokenCredential{}, fmt.Errorf("invalid tunnel token")
	}

	token, err := s.tokens.FindByPrefix(ctx, parts[0])

	if err != nil {
		return TunnelTokenCredential{}, fmt.Errorf("find tunnel token: %w", err)
	}

	now := s.now()

	if token.RevokedAt != nil || (token.ExpiresAt != nil && !token.ExpiresAt.After(now)) || !auth.EqualHash(token.TokenHash, parts[1]) {
		return TunnelTokenCredential{}, fmt.Errorf("invalid tunnel token")
	}

	if err := s.tokens.Touch(ctx, token.ID, now); err != nil {
		return TunnelTokenCredential{}, fmt.Errorf("touch tunnel token: %w", err)
	}

	var scopes []string

	if err := json.Unmarshal([]byte(token.Scopes), &scopes); err != nil {
		return TunnelTokenCredential{}, fmt.Errorf("decode tunnel token scopes: %w", err)
	}

	token.LastUsedAt = &now
	return TunnelTokenCredential{Token: token, Scopes: scopes}, nil
}

func (s *TunnelTokenService) Revoke(ctx context.Context, id string) error {

	if err := s.tokens.Revoke(ctx, id, s.now()); err != nil {
		return fmt.Errorf("revoke tunnel token: %w", err)
	}

	return nil
}
