package postgres

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func Migrate(db *gorm.DB) error {

	if db == nil {
		return fmt.Errorf("postgres database is required")
	}

	err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, name VARCHAR(200) NOT NULL, applied_at TIMESTAMPTZ NOT NULL)").Error

	if err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var applied []int64

		if err := tx.Raw("SELECT version FROM schema_migrations ORDER BY version").Scan(&applied).Error; err != nil {
			return fmt.Errorf("read schema migrations: %w", err)
		}

		appliedVersions := make(map[int64]struct{}, len(applied))

		for _, version := range applied {
			appliedVersions[version] = struct{}{}
		}

		for _, migration := range migrations() {

			if _, ok := appliedVersions[migration.version]; ok {
				continue
			}

			if err := migration.up(tx); err != nil {
				return fmt.Errorf("apply migration %d %s: %w", migration.version, migration.name, err)
			}

			if err := tx.Exec("INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)", migration.version, migration.name, time.Now().UTC()).Error; err != nil {
				return fmt.Errorf("record migration %d: %w", migration.version, err)
			}
		}

		return nil
	})
}

type migration struct {
	version int64
	name    string
	up      func(*gorm.DB) error
}

func migrations() []migration {
	return []migration{{version: 1, name: "initial_control_plane", up: func(db *gorm.DB) error {
		return db.AutoMigrate(
			&models.User{},
			&models.OAuthIdentity{},
			&models.Session{},
			&models.DeviceLogin{},
			&models.APIKey{},
			&models.Organization{},
			&models.OrganizationMember{},
			&models.Agent{},
			&models.Tunnel{},
			&models.TunnelToken{},
			&models.Domain{},
			&models.Plan{},
			&models.Subscription{},
			&models.BillingEvent{},
			&models.BillingCredential{},
			&models.UsageSnapshot{},
			&models.UsageEvent{},
			&models.AuditEvent{},
			&models.PlatformAdmin{},
			&models.OrganizationInvitation{},
		)
	}}, {version: 2, name: "plan_connection_limits", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Plan{})
	}}, {version: 3, name: "tunnel_passwords_and_request_usage", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Tunnel{}, &models.UsageEvent{})
	}}, {version: 4, name: "platform_admin_users", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.PlatformAdmin{})
	}}, {version: 5, name: "platform_admin_roles", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.PlatformAdmin{})
	}}, {version: 6, name: "platform_admin_names", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.PlatformAdmin{})
	}}, {version: 7, name: "organization_invitations", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.OrganizationInvitation{})
	}}, {version: 8, name: "usage_event_request_metadata", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.UsageEvent{})
	}}, {version: 9, name: "nullable_usage_client_ip", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.UsageEvent{})
	}}, {version: 10, name: "billing_invoices_and_receipts", up: func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Invoice{}, &models.Receipt{})
	}}}
}

func WithTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {

	if db == nil || fn == nil {
		return fmt.Errorf("database and transaction function are required")
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("run transaction: %w", err)
	}

	return nil
}
