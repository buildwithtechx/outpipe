package services

import (
	"context"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type BillingCredentialVault struct {
	repository repositories.BillingRepository
	protector  billingSecretProtector
	now        func() time.Time
}

func NewBillingCredentialVault(repository repositories.BillingRepository, protector billingSecretProtector) (*BillingCredentialVault, error) {

	if repository == nil || protector == nil {
		return nil, fmt.Errorf("billing repository and secret protector are required")
	}

	return &BillingCredentialVault{repository: repository, protector: protector, now: time.Now}, nil
}

func (v *BillingCredentialVault) Rotate(ctx context.Context, organizationID string, provider models.BillingProvider, kind, value string, expiresAt *time.Time) error {

	if organizationID == "" || provider == "" || kind == "" || value == "" {
		return fmt.Errorf("complete billing credential is required")
	}

	ciphertext, err := v.protector.Seal(value)

	if err != nil {
		return fmt.Errorf("encrypt billing credential: %w", err)
	}

	credential := models.BillingCredential{OrganizationID: organizationID, Provider: provider, Kind: kind, Ciphertext: ciphertext, RotatedAt: v.now(), ExpiresAt: expiresAt}

	if existing, findErr := v.repository.FindCredential(ctx, organizationID, provider, kind); findErr == nil {
		credential.ID = existing.ID
	}

	if err := v.repository.SaveCredential(ctx, &credential); err != nil {
		return fmt.Errorf("save rotated credential: %w", err)
	}

	return nil
}

func (v *BillingCredentialVault) Read(ctx context.Context, organizationID string, provider models.BillingProvider, kind string) (string, error) {
	credential, err := v.repository.FindCredential(ctx, organizationID, provider, kind)

	if err != nil {
		return "", fmt.Errorf("find billing credential: %w", err)
	}

	value, err := v.protector.Open(credential.Ciphertext)

	if err != nil {
		return "", fmt.Errorf("decrypt billing credential: %w", err)
	}

	return value, nil
}
