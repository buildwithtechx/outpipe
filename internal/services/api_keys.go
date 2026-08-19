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

type APIKeyService struct {
	keys repositories.APIKeyRepository
	now  func() time.Time
}

type APIKeyCredential struct {
	Key    models.APIKey
	Scopes []string
}

func NewAPIKeyService(keys repositories.APIKeyRepository) (*APIKeyService, error) {

	if keys == nil {
		return nil, fmt.Errorf("api key repository is required")
	}

	return &APIKeyService{keys: keys, now: time.Now}, nil
}

func (s *APIKeyService) Create(ctx context.Context, userID, name string, scopes []string, expiresAt *time.Time) (string, models.APIKey, error) {

	if userID == "" || strings.TrimSpace(name) == "" {
		return "", models.APIKey{}, fmt.Errorf("user id and name are required")
	}

	prefix, err := auth.NewToken("cdk", 8)

	if err != nil {
		return "", models.APIKey{}, err
	}

	secret, err := auth.NewToken("", 32)

	if err != nil {
		return "", models.APIKey{}, err
	}

	scopeJSON, err := json.Marshal(scopes)

	if err != nil {
		return "", models.APIKey{}, fmt.Errorf("encode api key scopes: %w", err)
	}

	key := models.APIKey{UserID: userID, Name: strings.TrimSpace(name), Prefix: prefix, SecretHash: auth.HashToken(secret), Scopes: string(scopeJSON), ExpiresAt: expiresAt}

	if err := s.keys.Create(ctx, &key); err != nil {
		return "", models.APIKey{}, fmt.Errorf("create api key: %w", err)
	}

	return prefix + "." + secret, key, nil
}

func (s *APIKeyService) CreateForOrganization(ctx context.Context, userID, organizationID, name string, scopes []string, expiresAt *time.Time) (string, models.APIKey, error) {

	if userID == "" || organizationID == "" {
		return "", models.APIKey{}, fmt.Errorf("user id and organization id are required")
	}

	prefix, err := auth.NewToken("cdk", 8)

	if err != nil {
		return "", models.APIKey{}, err
	}

	secret, err := auth.NewToken("", 32)

	if err != nil {
		return "", models.APIKey{}, err
	}

	scopeJSON, err := json.Marshal(scopes)

	if err != nil {
		return "", models.APIKey{}, fmt.Errorf("encode api key scopes: %w", err)
	}

	organizationID = strings.TrimSpace(organizationID)
	key := models.APIKey{UserID: userID, OrganizationID: &organizationID, Name: strings.TrimSpace(name), Prefix: prefix, SecretHash: auth.HashToken(secret), Scopes: string(scopeJSON), ExpiresAt: expiresAt}

	if err := s.keys.Create(ctx, &key); err != nil {
		return "", models.APIKey{}, fmt.Errorf("create api key: %w", err)
	}

	return prefix + "." + secret, key, nil
}

func (s *APIKeyService) List(ctx context.Context, userID string, organizationID *string) ([]models.APIKey, error) {

	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	keys, err := s.keys.ListByUser(ctx, userID, organizationID)

	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	return keys, nil
}

func (s *APIKeyService) RevokeOwned(ctx context.Context, userID string, organizationID *string, keyID string) error {

	if userID == "" || keyID == "" {
		return fmt.Errorf("user id and api key id are required")
	}

	key, err := s.keys.FindByID(ctx, keyID)

	if err != nil {
		return err
	}

	if key.UserID != userID {
		return fmt.Errorf("api key does not belong to this account")
	}

	if organizationID != nil && (key.OrganizationID == nil || *key.OrganizationID != *organizationID) {
		return fmt.Errorf("api key does not belong to this organization")
	}

	if organizationID == nil && key.OrganizationID != nil {
		return fmt.Errorf("api key does not belong to this account")
	}

	if err := s.keys.Revoke(ctx, keyID, s.now()); err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}

	return nil
}

func (s *APIKeyService) Authenticate(ctx context.Context, raw string) (APIKeyCredential, error) {
	parts := strings.SplitN(raw, ".", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return APIKeyCredential{}, fmt.Errorf("invalid api key")
	}

	key, err := s.keys.FindByPrefix(ctx, parts[0])

	if err != nil {
		return APIKeyCredential{}, fmt.Errorf("find api key: %w", err)
	}

	now := s.now()

	if key.RevokedAt != nil || (key.ExpiresAt != nil && !key.ExpiresAt.After(now)) || !auth.EqualHash(key.SecretHash, parts[1]) {
		return APIKeyCredential{}, fmt.Errorf("invalid api key")
	}

	if err := s.keys.Touch(ctx, key.ID, now); err != nil {
		return APIKeyCredential{}, fmt.Errorf("touch api key: %w", err)
	}

	var scopes []string

	if err := json.Unmarshal([]byte(key.Scopes), &scopes); err != nil {
		return APIKeyCredential{}, fmt.Errorf("decode api key scopes: %w", err)
	}

	key.LastUsedAt = &now
	return APIKeyCredential{Key: key, Scopes: scopes}, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, id string) error {

	if err := s.keys.Revoke(ctx, id, s.now()); err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}

	return nil
}
