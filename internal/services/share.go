package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type ShareService struct {
	repo repositories.SecretRepository
}

func NewShareService(repo repositories.SecretRepository) (*ShareService, error) {
	if repo == nil {
		return nil, fmt.Errorf("secret repository is required")
	}
	return &ShareService{repo: repo}, nil
}

type CreateShareInput struct {
	OrganizationID   *string `json:"organizationId,omitempty"`
	ProjectID        *string `json:"projectId,omitempty"`
	EnvironmentID    *string `json:"environmentId,omitempty"`
	CreatedByID      *string `json:"createdById,omitempty"`
	Ciphertext       string  `json:"ciphertext"`
	IV               string  `json:"iv"`
	KeyVerifier      string  `json:"keyVerifier"`
	PasswordSalt     *string `json:"passwordSalt,omitempty"`
	PasswordVerifier *string `json:"passwordVerifier,omitempty"`
	ContentFormat    string  `json:"contentFormat"`
	TTLSeconds       int     `json:"ttlSeconds"`
	MaxViews         int     `json:"maxViews"`
}

func (s *ShareService) CreateShare(ctx context.Context, input CreateShareInput) (models.SecretShareLink, error) {
	if input.OrganizationID != nil {
		project, environment := "", ""
		if input.ProjectID != nil {
			project = *input.ProjectID
		}
		if input.EnvironmentID != nil {
			environment = *input.EnvironmentID
		}
		project, environment, err := resolveSecretScope(ctx, s.repo, *input.OrganizationID, project, environment)
		if err != nil {
			return models.SecretShareLink{}, fmt.Errorf("authorize share scope: %w", err)
		}
		if project != "" {
			input.ProjectID = &project
		}
		if environment != "" {
			input.EnvironmentID = &environment
		}
	}
	if input.Ciphertext == "" || input.IV == "" || input.KeyVerifier == "" {
		return models.SecretShareLink{}, fmt.Errorf("ciphertext, iv, and keyVerifier are required")
	}
	if input.MaxViews <= 0 || input.MaxViews > 100 {
		input.MaxViews = 1
	}
	if input.TTLSeconds <= 0 || input.TTLSeconds > 7*86400 {
		input.TTLSeconds = 86400
	}
	if input.ContentFormat == "" {
		input.ContentFormat = "bundle"
	}

	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return models.SecretShareLink{}, fmt.Errorf("generate share id: %w", err)
	}

	link := models.SecretShareLink{
		ID:               hex.EncodeToString(randomID),
		OrganizationID:   input.OrganizationID,
		ProjectID:        input.ProjectID,
		EnvironmentID:    input.EnvironmentID,
		CreatedByID:      input.CreatedByID,
		Ciphertext:       input.Ciphertext,
		IV:               input.IV,
		KeyVerifier:      input.KeyVerifier,
		PasswordSalt:     input.PasswordSalt,
		PasswordVerifier: input.PasswordVerifier,
		ContentFormat:    input.ContentFormat,
		ExpiresAt:        time.Now().UTC().Add(time.Duration(input.TTLSeconds) * time.Second),
		MaxViews:         input.MaxViews,
		Views:            0,
	}

	if err := s.repo.CreateShareLink(ctx, &link); err != nil {
		return models.SecretShareLink{}, fmt.Errorf("create share link: %w", err)
	}
	return link, nil
}

type ShareMetaDTO struct {
	ID             string    `json:"id"`
	ContentFormat  string    `json:"contentFormat"`
	ExpiresAt      time.Time `json:"expiresAt"`
	MaxViews       int       `json:"maxViews"`
	ViewsRemaining int       `json:"viewsRemaining"`
	NeedsPassword  bool      `json:"needsPassword"`
}

func (s *ShareService) GetShareMeta(ctx context.Context, id string) (ShareMetaDTO, error) {
	link, err := s.repo.FindShareLink(ctx, id)
	if err != nil {
		return ShareMetaDTO{}, err
	}
	remaining := link.MaxViews - link.Views
	if remaining < 0 {
		remaining = 0
	}
	return ShareMetaDTO{
		ID:             link.ID,
		ContentFormat:  link.ContentFormat,
		ExpiresAt:      link.ExpiresAt,
		MaxViews:       link.MaxViews,
		ViewsRemaining: remaining,
		NeedsPassword:  link.PasswordSalt != nil && *link.PasswordSalt != "",
	}, nil
}

type RevealShareResult struct {
	Ciphertext    string `json:"ciphertext"`
	IV            string `json:"iv"`
	ContentFormat string `json:"contentFormat"`
}

func (s *ShareService) RevealShare(ctx context.Context, id, keyVerifier string, passwordVerifier *string) (RevealShareResult, error) {
	link, err := s.repo.FindShareLink(ctx, id)
	if err != nil {
		return RevealShareResult{}, err
	}
	if link.KeyVerifier != keyVerifier {
		return RevealShareResult{}, errors.New("invalid key verifier")
	}
	if link.PasswordVerifier != nil && *link.PasswordVerifier != "" {
		if passwordVerifier == nil || *passwordVerifier != *link.PasswordVerifier {
			return RevealShareResult{}, errors.New("invalid password")
		}
	}

	revealed, err := s.repo.RevealShareLink(ctx, id, time.Now().UTC())
	if err != nil {
		return RevealShareResult{}, fmt.Errorf("reveal share: %w", err)
	}

	return RevealShareResult{
		Ciphertext:    revealed.Ciphertext,
		IV:            revealed.IV,
		ContentFormat: revealed.ContentFormat,
	}, nil
}

type OrgShareDTO struct {
	ID            string     `json:"id"`
	ProjectID     *string    `json:"projectId,omitempty"`
	EnvironmentID *string    `json:"environmentId,omitempty"`
	CreatedByID   *string    `json:"createdById,omitempty"`
	ContentFormat string     `json:"contentFormat"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	MaxViews      int        `json:"maxViews"`
	Views         int        `json:"views"`
	RevokedAt     *time.Time `json:"revokedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (s *ShareService) ListOrgShares(ctx context.Context, orgID string) ([]OrgShareDTO, error) {
	links, err := s.repo.ListOrgShares(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list share metadata: %w", err)
	}
	items := make([]OrgShareDTO, 0, len(links))
	for _, link := range links {
		items = append(items, OrgShareDTO{ID: link.ID, ProjectID: link.ProjectID, EnvironmentID: link.EnvironmentID, CreatedByID: link.CreatedByID, ContentFormat: link.ContentFormat, ExpiresAt: link.ExpiresAt, MaxViews: link.MaxViews, Views: link.Views, RevokedAt: link.RevokedAt, CreatedAt: link.CreatedAt})
	}
	return items, nil
}

func (s *ShareService) RevokeShare(ctx context.Context, id string) error {
	return s.repo.RevokeShareLink(ctx, id)
}

func (s *ShareService) RevokeOrgShare(ctx context.Context, orgID, id string) error {
	link, err := s.repo.FindShareLink(ctx, id)
	if err != nil {
		return fmt.Errorf("find organization share: %w", err)
	}
	if link.OrganizationID == nil || *link.OrganizationID != orgID {
		return repositories.ErrNotFound
	}
	return s.repo.RevokeShareLink(ctx, id)
}
