package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/security"
)

type SecretService struct {
	repo          repositories.SecretRepository
	encryptionKey []byte
}

func NewSecretService(repo repositories.SecretRepository, encryptionKey string) (*SecretService, error) {
	if repo == nil {
		return nil, fmt.Errorf("secret repository is required")
	}
	if strings.TrimSpace(encryptionKey) == "" {
		return nil, fmt.Errorf("secret encryption key is required")
	}
	keyBytes := parseEncryptionKey(encryptionKey)
	return &SecretService{
		repo:          repo,
		encryptionKey: keyBytes,
	}, nil
}

func parseEncryptionKey(raw string) []byte {
	decoded, err := hex.DecodeString(strings.TrimSpace(raw))
	if err == nil && len(decoded) == 32 {
		return decoded
	}
	h := sha256.Sum256([]byte(raw))
	return h[:]
}

func (s *SecretService) CreateProject(ctx context.Context, orgID, slug, name, description, userID string) (models.SecretProject, error) {
	if err := validateSecretSlug(slug, 120); err != nil {
		return models.SecretProject{}, err
	}
	project := models.SecretProject{
		OrganizationID: orgID,
		Slug:           strings.ToLower(strings.TrimSpace(slug)),
		Name:           strings.TrimSpace(name),
		Description:    strings.TrimSpace(description),
		CreatedByID:    userID,
	}
	if err := s.repo.CreateProject(ctx, &project); err != nil {
		return models.SecretProject{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (s *SecretService) ListProjects(ctx context.Context, orgID string) ([]models.SecretProject, error) {
	return s.repo.ListProjects(ctx, orgID)
}

func (s *SecretService) GetProject(ctx context.Context, orgID, slug string) (models.SecretProject, error) {
	return s.repo.FindProjectBySlug(ctx, orgID, slug)
}

func (s *SecretService) DeleteProject(ctx context.Context, orgID, projectID string) error {
	return s.repo.DeleteProject(ctx, orgID, projectID)
}

func (s *SecretService) CreateEnvironment(ctx context.Context, orgID, projectID, slug, name, userID string) (models.SecretEnvironment, error) {
	if err := validateSecretSlug(slug, 64); err != nil {
		return models.SecretEnvironment{}, err
	}
	projectID, _, err := s.ResolveScope(ctx, orgID, projectID, "")
	if err != nil {
		return models.SecretEnvironment{}, err
	}
	env := models.SecretEnvironment{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Slug:           strings.ToLower(strings.TrimSpace(slug)),
		Name:           strings.TrimSpace(name),
		CreatedByID:    userID,
	}
	if err := s.repo.CreateEnvironment(ctx, &env); err != nil {
		return models.SecretEnvironment{}, fmt.Errorf("create environment: %w", err)
	}
	return env, nil
}

func (s *SecretService) ListEnvironments(ctx context.Context, orgID, projectID string) ([]models.SecretEnvironment, error) {
	projectID, _, err := s.ResolveScope(ctx, orgID, projectID, "")
	if err != nil {
		return nil, err
	}
	return s.repo.ListEnvironments(ctx, orgID, projectID)
}

type SecretItemDTO struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	Comment   string    `json:"comment"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *SecretService) SetSecret(ctx context.Context, orgID, projectID, envID, key, value, comment, userID string) (SecretItemDTO, error) {
	projectID, envID, err := s.ResolveScope(ctx, orgID, projectID, envID)
	if err != nil {
		return SecretItemDTO{}, err
	}
	key = strings.TrimSpace(key)
	ciphertext, iv, err := security.EncryptAESGCM([]byte(value), s.encryptionKey)
	if err != nil {
		return SecretItemDTO{}, fmt.Errorf("encrypt secret: %w", err)
	}
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	encodedIV := base64.StdEncoding.EncodeToString(iv)

	entries, err := s.repo.ListEntries(ctx, orgID, projectID, envID, false)
	if err != nil {
		return SecretItemDTO{}, err
	}

	var existing *models.SecretEntry
	for i := range entries {
		if entries[i].Key == key {
			existing = &entries[i]
			break
		}
	}

	if existing == nil {
		entry := models.SecretEntry{
			OrganizationID: orgID,
			ProjectID:      projectID,
			EnvironmentID:  envID,
			Key:            key,
			Comment:        comment,
		}
		ver := models.SecretVersion{
			Ciphertext:  encodedCiphertext,
			IV:          encodedIV,
			KeyVersion:  1,
			CreatedByID: userID,
		}
		if err := s.repo.CreateEntryWithVersion(ctx, &entry, &ver); err != nil {
			return SecretItemDTO{}, fmt.Errorf("create secret: %w", err)
		}
		return SecretItemDTO{
			ID:        entry.ID,
			Key:       entry.Key,
			Value:     value,
			Comment:   entry.Comment,
			Version:   ver.Version,
			UpdatedAt: entry.UpdatedAt,
		}, nil
	}

	ver := models.SecretVersion{
		Ciphertext:  encodedCiphertext,
		IV:          encodedIV,
		KeyVersion:  1,
		CreatedByID: userID,
	}
	if err := s.repo.UpdateEntryVersion(ctx, orgID, existing.ID, &ver); err != nil {
		return SecretItemDTO{}, fmt.Errorf("update secret version: %w", err)
	}
	return SecretItemDTO{
		ID:        existing.ID,
		Key:       existing.Key,
		Value:     value,
		Comment:   existing.Comment,
		Version:   ver.Version,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (s *SecretService) ListSecrets(ctx context.Context, orgID, projectID, envID string, reveal bool) ([]SecretItemDTO, error) {
	projectID, envID, err := s.ResolveScope(ctx, orgID, projectID, envID)
	if err != nil {
		return nil, err
	}
	entries, err := s.repo.ListEntries(ctx, orgID, projectID, envID, false)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	items := make([]SecretItemDTO, 0, len(entries))
	for _, entry := range entries {
		ver, err := s.repo.FindLatestVersion(ctx, orgID, entry.ID)
		if err != nil {
			return nil, fmt.Errorf("read secret version: %w", err)
		}
		val := ""
		if reveal {
			rawCipher, err := base64.StdEncoding.DecodeString(ver.Ciphertext)
			if err != nil {
				return nil, fmt.Errorf("decode secret ciphertext: %w", err)
			}
			rawIV, err := base64.StdEncoding.DecodeString(ver.IV)
			if err != nil {
				return nil, fmt.Errorf("decode secret nonce: %w", err)
			}
			decrypted, err := security.DecryptAESGCM(rawCipher, rawIV, s.encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("decrypt secret value: %w", err)
			}
			val = string(decrypted)
		}
		items = append(items, SecretItemDTO{
			ID:        entry.ID,
			Key:       entry.Key,
			Value:     val,
			Comment:   entry.Comment,
			Version:   ver.Version,
			UpdatedAt: entry.UpdatedAt,
		})
	}
	return items, nil
}

func (s *SecretService) SoftDeleteSecret(ctx context.Context, orgID, entryID string) error {
	return s.repo.SoftDeleteEntry(ctx, orgID, entryID, nil)
}

func (s *SecretService) RestoreSecret(ctx context.Context, orgID, entryID string) error {
	return s.repo.RestoreEntry(ctx, orgID, entryID)
}

func (s *SecretService) ListTrash(ctx context.Context, orgID, projectID, envID string) ([]models.SecretEntry, error) {
	projectID, envID, err := s.ResolveScope(ctx, orgID, projectID, envID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListEntries(ctx, orgID, projectID, envID, true)
}
