package repositories

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func appendSecretVersion(tx *gorm.DB, orgID, entryID string, version *models.SecretVersion) error {
	result := tx.Model(&models.SecretEntry{}).Where("organization_id = ? AND id = ? AND deleted_at IS NULL", orgID, entryID).Update("updated_at", time.Now().UTC())
	if result.Error != nil {
		return fmt.Errorf("lock secret entry: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	var lastVersion int
	if err := tx.Model(&models.SecretVersion{}).Where("organization_id = ? AND entry_id = ?", orgID, entryID).Select("COALESCE(MAX(version), 0)").Row().Scan(&lastVersion); err != nil {
		return fmt.Errorf("query max secret version: %w", err)
	}
	version.EntryID, version.OrganizationID, version.Version = entryID, orgID, lastVersion+1
	return wrap(tx.Create(version).Error, "insert secret version")
}
