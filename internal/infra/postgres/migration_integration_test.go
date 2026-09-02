package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMigrationRunnerIsRestartSafe(t *testing.T) {
	dsn := os.Getenv("OUTPIPE_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("OUTPIPE_MIGRATION_TEST_DSN is not configured")
	}

	db, err := Open(context.Background(), Config{DSN: dsn, MaxOpenConns: 2, MaxIdleConns: 2, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get migration database: %v", err)
	}
	defer sqlDB.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("run clean migration: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("rerun upgraded migration: %v", err)
	}

	var applied int64
	if err := db.Table("schema_migrations").Count(&applied).Error; err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if applied != int64(len(migrations())) {
		t.Fatalf("applied migrations = %d, want %d", applied, len(migrations()))
	}
}
