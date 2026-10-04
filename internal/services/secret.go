package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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
			Version:   1,
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
	entries, err := s.repo.ListEntries(ctx, orgID, projectID, envID, false)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	items := make([]SecretItemDTO, 0, len(entries))
	for _, entry := range entries {
		ver, err := s.repo.FindLatestVersion(ctx, orgID, entry.ID)
		if err != nil {
			continue
		}
		val := ""
		if reveal {
			rawCipher, err1 := base64.StdEncoding.DecodeString(ver.Ciphertext)
			rawIV, err2 := base64.StdEncoding.DecodeString(ver.IV)
			if err1 == nil && err2 == nil {
				decrypted, err := security.DecryptAESGCM(rawCipher, rawIV, s.encryptionKey)
				if err == nil {
					val = string(decrypted)
				}
			}
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
	return s.repo.ListEntries(ctx, orgID, projectID, envID, true)
}

type CreatedMachineTokenDTO struct {
	Token models.SecretMachineToken `json:"token"`
	Raw   string                    `json:"raw"`
}

func (s *SecretService) CreateMachineToken(ctx context.Context, orgID, projectID, envID, name string, scopes []string, userID string) (CreatedMachineTokenDTO, error) {
	entropy := make([]byte, 24)
	if _, err := rand.Read(entropy); err != nil {
		return CreatedMachineTokenDTO{}, fmt.Errorf("generate entropy: %w", err)
	}
	prefix := "op_sec_" + hex.EncodeToString(entropy[:4])
	raw := prefix + "_" + hex.EncodeToString(entropy[4:])
	tokenHash := security.SignHMACSHA256([]byte(raw), hex.EncodeToString(s.encryptionKey))

	scopesJSON, _ := json.Marshal(scopes)
	token := models.SecretMachineToken{
		OrganizationID: orgID,
		Name:           name,
		Prefix:         prefix,
		TokenHash:      tokenHash,
		Scopes:         string(scopesJSON),
		CreatedByID:    userID,
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
	return s.repo.FindMachineTokenByHash(ctx, tokenHash)
}

func (s *SecretService) ListMachineTokens(ctx context.Context, orgID string) ([]models.SecretMachineToken, error) {
	return s.repo.ListMachineTokens(ctx, orgID)
}

func (s *SecretService) RevokeMachineToken(ctx context.Context, orgID, tokenID string) error {
	return s.repo.RevokeMachineToken(ctx, orgID, tokenID)
}
