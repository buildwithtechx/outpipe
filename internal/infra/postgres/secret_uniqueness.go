package postgres

import (
	"fmt"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func migrateSecretUniqueness(db *gorm.DB) error {
	if err := db.Exec(`WITH ranked AS (
		SELECT id, row_number() OVER (PARTITION BY organization_id, project_id, environment_id, key ORDER BY updated_at DESC, id DESC) AS position
		FROM secret_entries WHERE deleted_at IS NULL
	) UPDATE secret_entries SET deleted_at = CURRENT_TIMESTAMP WHERE id IN (SELECT id FROM ranked WHERE position > 1)`).Error; err != nil {
		return fmt.Errorf("archive duplicate active secret entries: %w", err)
	}
	if err := db.Exec(`WITH ranked AS (
		SELECT id, row_number() OVER (PARTITION BY entry_id ORDER BY version, created_at, id) AS position
		FROM secret_versions WHERE entry_id IN (SELECT entry_id FROM secret_versions GROUP BY entry_id, version HAVING COUNT(*) > 1)
	) UPDATE secret_versions SET version = ranked.position FROM ranked WHERE secret_versions.id = ranked.id`).Error; err != nil {
		return fmt.Errorf("normalize secret version numbers: %w", err)
	}
	return db.AutoMigrate(&models.SecretEntry{}, &models.SecretVersion{})
}
