package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
)

type CreatedMachineTokenDTO struct {
	Token models.SecretMachineToken `json:"token"`
	Raw   string                    `json:"raw"`
}

func (s *SecretService) CreateMachineToken(ctx context.Context, orgID, projectID, envID, name string, scopes []string, userID string) (CreatedMachineTokenDTO, error) {
	projectID, envID, err := s.ResolveScope(ctx, orgID, projectID, envID)
	if err != nil {
		return CreatedMachineTokenDTO{}, err
	}
	entropy := make([]byte, 24)
	if _, err := rand.Read(entropy); err != nil {
		return CreatedMachineTokenDTO{}, fmt.Errorf("generate entropy: %w", err)
	}
	prefix := "op_sec_" + hex.EncodeToString(entropy[:4])
	raw := prefix + "_" + hex.EncodeToString(entropy[4:])
	tokenHash := security.SignHMACSHA256([]byte(raw), hex.EncodeToString(s.encryptionKey))

	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return CreatedMachineTokenDTO{}, fmt.Errorf("encode machine token scopes: %w", err)
	}
	token := models.SecretMachineToken{
		OrganizationID: orgID,
		Name:           name,
		Prefix:         prefix,
		TokenHash:      tokenHash,
		Scopes:         string(scopesJSON),
		CreatedByID:    userID,
		ExpiresAt:      func() *time.Time { value := time.Now().UTC().Add(24 * time.Hour); return &value }(),
	}
	if projectID != "" {
		token.ProjectID = &projectID
	}
	if envID != "" {
		token.EnvironmentID = &envID
	}
	if err := s.repo.CreateMachineToken(ctx, &token); err != nil {
		return CreatedMachineTokenDTO{}, fmt.Errorf("create token: %w", err)
	}
	return CreatedMachineTokenDTO{Token: token, Raw: raw}, nil
}

func (s *SecretService) VerifyMachineToken(ctx context.Context, raw string) (models.SecretMachineToken, error) {
	tokenHash := security.SignHMACSHA256([]byte(raw), hex.EncodeToString(s.encryptionKey))
	token, err := s.repo.FindMachineTokenByHash(ctx, tokenHash)
	if err != nil {
		return models.SecretMachineToken{}, fmt.Errorf("find machine token: %w", err)
	}
	if token.RevokedAt != nil || token.ExpiresAt == nil || !token.ExpiresAt.After(time.Now()) || token.ExpiresAt.After(time.Now().Add(24*time.Hour)) {
		return models.SecretMachineToken{}, fmt.Errorf("machine token is expired or lacks a short lifetime")
	}
	return token, nil
}

func (s *SecretService) ListMachineTokens(ctx context.Context, orgID string) ([]models.SecretMachineToken, error) {
	return s.repo.ListMachineTokens(ctx, orgID)
}

func (s *SecretService) RevokeMachineToken(ctx context.Context, orgID, tokenID string) error {
	return s.repo.RevokeMachineToken(ctx, orgID, tokenID)
}
