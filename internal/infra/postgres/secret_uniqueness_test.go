package postgres

import (
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
)

func TestSecretUniquenessUpgradePreservesHistory(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	}()
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.SecretEntry{}, &models.SecretVersion{}); err != nil {
		t.Fatal(err)
	}
	for _, index := range []string{"idx_secret_active_key", "idx_secret_entry_version"} {
		if err := db.Exec("DROP INDEX " + index).Error; err != nil {
			t.Fatal(err)
		}
	}
	first := models.SecretEntry{OrganizationID: "org", ProjectID: "project", EnvironmentID: "environment", Key: "KEY"}
	second := first
	if err := db.Create(&first).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	for _, entry := range []models.SecretEntry{first, first, second} {
		version := models.SecretVersion{OrganizationID: "org", EntryID: entry.ID, Version: 1, Ciphertext: "preserved", IV: "nonce"}
		if err := db.Create(&version).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := migrateSecretUniqueness(db); err != nil {
		t.Fatal(err)
	}
	if err := migrateSecretUniqueness(db); err != nil {
		t.Fatal(err)
	}
	var total, active, versions int64
	if err := db.Model(&models.SecretEntry{}).Count(&total).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.SecretEntry{}).Where("deleted_at IS NULL").Count(&active).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.SecretVersion{}).Count(&versions).Error; err != nil {
		t.Fatal(err)
	}
	if total != 2 || active != 1 || versions != 3 {
		t.Fatalf("lost history or retained duplicates: total=%d active=%d versions=%d", total, active, versions)
	}
	third := models.SecretEntry{OrganizationID: "org", ProjectID: "project", EnvironmentID: "environment", Key: "KEY"}
	if err := db.Create(&third).Error; err == nil {
		t.Fatal("active secret uniqueness is not enforced")
	}
	duplicate := models.SecretVersion{OrganizationID: "org", EntryID: first.ID, Version: 1, Ciphertext: "duplicate", IV: "nonce"}
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("version uniqueness is not enforced")
	}
}
