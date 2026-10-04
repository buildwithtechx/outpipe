package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

type SecretRepository interface {
	CreateProject(context.Context, *models.SecretProject) error
	FindProjectBySlug(context.Context, string, string) (models.SecretProject, error)
	ListProjects(context.Context, string) ([]models.SecretProject, error)
	DeleteProject(context.Context, string, string) error

	CreateEnvironment(context.Context, *models.SecretEnvironment) error
	FindEnvironmentBySlug(context.Context, string, string, string) (models.SecretEnvironment, error)
	ListEnvironments(context.Context, string, string) ([]models.SecretEnvironment, error)

	CreateEntryWithVersion(context.Context, *models.SecretEntry, *models.SecretVersion) error
	UpdateEntryVersion(context.Context, string, string, *models.SecretVersion) error
	ListEntries(context.Context, string, string, string, bool) ([]models.SecretEntry, error)
	FindEntry(context.Context, string, string) (models.SecretEntry, error)
	FindLatestVersion(context.Context, string, string) (models.SecretVersion, error)
	SoftDeleteEntry(context.Context, string, string, *string) error
	RestoreEntry(context.Context, string, string) error
	PermanentDeleteEntry(context.Context, string, string) error

	CreateMachineToken(context.Context, *models.SecretMachineToken) error
	FindMachineTokenByHash(context.Context, string) (models.SecretMachineToken, error)
	ListMachineTokens(context.Context, string) ([]models.SecretMachineToken, error)
	RevokeMachineToken(context.Context, string, string) error

	RecordAuditEvent(context.Context, *models.SecretAuditEvent) error
	ListAuditEvents(context.Context, string, int) ([]models.SecretAuditEvent, error)

	CreateShareLink(context.Context, *models.SecretShareLink) error
	FindShareLink(context.Context, string) (models.SecretShareLink, error)
	RevealShareLink(context.Context, string, time.Time) (models.SecretShareLink, error)
	RevokeShareLink(context.Context, string) error
	ListOrgShares(context.Context, string) ([]models.SecretShareLink, error)
}

type GormSecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) (*GormSecretRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &GormSecretRepository{db: db}, nil
}

func (r *GormSecretRepository) CreateProject(ctx context.Context, p *models.SecretProject) error {
	if p == nil {
		return fmt.Errorf("secret project is required")
	}
	return wrap(r.db.WithContext(ctx).Create(p).Error, "create secret project")
}

func (r *GormSecretRepository) FindProjectBySlug(ctx context.Context, orgID, slug string) (models.SecretProject, error) {
	var project models.SecretProject
	err := r.db.WithContext(ctx).Where("organization_id = ? AND slug = ?", orgID, slug).First(&project).Error
	if err != nil {
		return models.SecretProject{}, mapError(err)
	}
	return project, nil
}

func (r *GormSecretRepository) ListProjects(ctx context.Context, orgID string) ([]models.SecretProject, error) {
	var projects []models.SecretProject
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at asc").Find(&projects).Error
	return projects, wrap(err, "list secret projects")
}

func (r *GormSecretRepository) DeleteProject(ctx context.Context, orgID, projectID string) error {
	return wrap(r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", orgID, projectID).Delete(&models.SecretProject{}).Error, "delete secret project")
}

func (r *GormSecretRepository) CreateEnvironment(ctx context.Context, env *models.SecretEnvironment) error {
	if env == nil {
		return fmt.Errorf("secret environment is required")
	}
	return wrap(r.db.WithContext(ctx).Create(env).Error, "create secret environment")
}

func (r *GormSecretRepository) FindEnvironmentBySlug(ctx context.Context, orgID, projectID, slug string) (models.SecretEnvironment, error) {
	var env models.SecretEnvironment
	err := r.db.WithContext(ctx).Where("organization_id = ? AND project_id = ? AND slug = ?", orgID, projectID, slug).First(&env).Error
	if err != nil {
		return models.SecretEnvironment{}, mapError(err)
	}
	return env, nil
}

func (r *GormSecretRepository) ListEnvironments(ctx context.Context, orgID, projectID string) ([]models.SecretEnvironment, error) {
	var envs []models.SecretEnvironment
	err := r.db.WithContext(ctx).Where("organization_id = ? AND project_id = ?", orgID, projectID).Order("created_at asc").Find(&envs).Error
	return envs, wrap(err, "list secret environments")
}

func (r *GormSecretRepository) CreateEntryWithVersion(ctx context.Context, entry *models.SecretEntry, version *models.SecretVersion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entry).Error; err != nil {
			return fmt.Errorf("create secret entry: %w", err)
		}
		version.EntryID = entry.ID
		version.OrganizationID = entry.OrganizationID
		version.Version = 1
		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("create secret version: %w", err)
		}
		return nil
	})
}

func (r *GormSecretRepository) UpdateEntryVersion(ctx context.Context, orgID, entryID string, version *models.SecretVersion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lastVersion int
		row := tx.Model(&models.SecretVersion{}).Where("organization_id = ? AND entry_id = ?", orgID, entryID).Select("COALESCE(MAX(version), 0)").Row()
		if err := row.Scan(&lastVersion); err != nil {
			return fmt.Errorf("query max version: %w", err)
		}
		version.EntryID = entryID
		version.OrganizationID = orgID
		version.Version = lastVersion + 1
		if err := tx.Create(version).Error; err != nil {
			return fmt.Errorf("insert secret version: %w", err)
		}
		return tx.Model(&models.SecretEntry{}).Where("organization_id = ? AND id = ?", orgID, entryID).Update("updated_at", time.Now().UTC()).Error
	})
}

func (r *GormSecretRepository) ListEntries(ctx context.Context, orgID, projectID, envID string, includeDeleted bool) ([]models.SecretEntry, error) {
	var entries []models.SecretEntry
	q := r.db.WithContext(ctx).Where("organization_id = ? AND project_id = ? AND environment_id = ?", orgID, projectID, envID)
	if !includeDeleted {
		q = q.Where("deleted_at IS NULL")
	} else {
		q = q.Where("deleted_at IS NOT NULL")
	}
	err := q.Order("key asc").Find(&entries).Error
	return entries, wrap(err, "list secret entries")
}

func (r *GormSecretRepository) FindEntry(ctx context.Context, orgID, entryID string) (models.SecretEntry, error) {
	var entry models.SecretEntry
	err := r.db.WithContext(ctx).Where("organization_id = ? AND id = ?", orgID, entryID).First(&entry).Error
	if err != nil {
		return models.SecretEntry{}, mapError(err)
	}
	return entry, nil
}

func (r *GormSecretRepository) FindLatestVersion(ctx context.Context, orgID, entryID string) (models.SecretVersion, error) {
	var version models.SecretVersion
	err := r.db.WithContext(ctx).Where("organization_id = ? AND entry_id = ?", orgID, entryID).Order("version desc").First(&version).Error
	if err != nil {
		return models.SecretVersion{}, mapError(err)
	}
	return version, nil
}

func (r *GormSecretRepository) SoftDeleteEntry(ctx context.Context, orgID, entryID string, batchID *string) error {
	now := time.Now().UTC()
	updates := map[string]any{"deleted_at": &now, "deletion_batch_id": batchID}
	return wrap(r.db.WithContext(ctx).Model(&models.SecretEntry{}).Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, entryID).Updates(updates).Error, "soft delete secret entry")
}

func (r *GormSecretRepository) RestoreEntry(ctx context.Context, orgID, entryID string) error {
	updates := map[string]any{"deleted_at": nil, "deletion_batch_id": nil}
	return wrap(r.db.WithContext(ctx).Model(&models.SecretEntry{}).Where("organization_id = ? AND id = ? AND deleted_at IS NOT NULL", orgID, entryID).Updates(updates).Error, "restore secret entry")
}

func (r *GormSecretRepository) PermanentDeleteEntry(ctx context.Context, orgID, entryID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ? AND entry_id = ?", orgID, entryID).Delete(&models.SecretVersion{}).Error; err != nil {
			return fmt.Errorf("delete secret versions: %w", err)
		}
		return tx.Where("organization_id = ? AND id = ?", orgID, entryID).Delete(&models.SecretEntry{}).Error
	})
}

func (r *GormSecretRepository) CreateMachineToken(ctx context.Context, token *models.SecretMachineToken) error {
	if token == nil {
		return fmt.Errorf("machine token is required")
	}
	return wrap(r.db.WithContext(ctx).Create(token).Error, "create machine token")
}

func (r *GormSecretRepository) FindMachineTokenByHash(ctx context.Context, tokenHash string) (models.SecretMachineToken, error) {
	var token models.SecretMachineToken
	err := r.db.WithContext(ctx).Where("token_hash = ? AND revoked_at IS NULL", tokenHash).First(&token).Error
	if err != nil {
		return models.SecretMachineToken{}, mapError(err)
	}
	return token, nil
}

func (r *GormSecretRepository) ListMachineTokens(ctx context.Context, orgID string) ([]models.SecretMachineToken, error) {
	var tokens []models.SecretMachineToken
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at desc").Find(&tokens).Error
	return tokens, wrap(err, "list machine tokens")
}

func (r *GormSecretRepository) RevokeMachineToken(ctx context.Context, orgID, tokenID string) error {
	now := time.Now().UTC()
	return wrap(r.db.WithContext(ctx).Model(&models.SecretMachineToken{}).Where("organization_id = ? AND id = ?", orgID, tokenID).Update("revoked_at", &now).Error, "revoke machine token")
}

func (r *GormSecretRepository) RecordAuditEvent(ctx context.Context, event *models.SecretAuditEvent) error {
	if event == nil {
		return fmt.Errorf("audit event is required")
	}
	return wrap(r.db.WithContext(ctx).Create(event).Error, "record secret audit event")
}

func (r *GormSecretRepository) ListAuditEvents(ctx context.Context, orgID string, limit int) ([]models.SecretAuditEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var events []models.SecretAuditEvent
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at desc").Limit(limit).Find(&events).Error
	return events, wrap(err, "list secret audit events")
}

func (r *GormSecretRepository) CreateShareLink(ctx context.Context, link *models.SecretShareLink) error {
	if link == nil {
		return fmt.Errorf("share link is required")
	}
	link.CreatedAt = time.Now().UTC()
	return wrap(r.db.WithContext(ctx).Create(link).Error, "create share link")
}

func (r *GormSecretRepository) FindShareLink(ctx context.Context, id string) (models.SecretShareLink, error) {
	var link models.SecretShareLink
	err := r.db.WithContext(ctx).Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, time.Now().UTC()).First(&link).Error
	if err != nil {
		return models.SecretShareLink{}, mapError(err)
	}
	return link, nil
}

func (r *GormSecretRepository) RevealShareLink(ctx context.Context, id string, now time.Time) (models.SecretShareLink, error) {
	var link models.SecretShareLink
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, now).First(&link).Error; err != nil {
			return err
		}
		if link.Views >= link.MaxViews {
			return ErrNotFound
		}
		link.Views++
		link.LastRevealedAt = &now
		return tx.Model(&models.SecretShareLink{}).Where("id = ?", id).Updates(map[string]any{
			"views":            link.Views,
			"last_revealed_at": link.LastRevealedAt,
		}).Error
	})
	if err != nil {
		return models.SecretShareLink{}, mapError(err)
	}
	return link, nil
}

func (r *GormSecretRepository) RevokeShareLink(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return wrap(r.db.WithContext(ctx).Model(&models.SecretShareLink{}).Where("id = ?", id).Update("revoked_at", &now).Error, "revoke share link")
}

func (r *GormSecretRepository) ListOrgShares(ctx context.Context, orgID string) ([]models.SecretShareLink, error) {
	var shares []models.SecretShareLink
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Order("created_at desc").Find(&shares).Error
	return shares, wrap(err, "list organization shares")
}
