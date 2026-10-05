package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
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
	schema := "migration_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	db = db.Session(&gorm.Session{NewDB: true})
	connection, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close isolated migration connection: %v", err)
		}
	}()
	defer func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	}()
	db.Statement.ConnPool = connection
	if err := db.Exec("SET search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}

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
	repo, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	past := now.Add(-time.Minute)
	monitor := models.UptimeMonitor{ID: "migration-uptime-check", OrganizationID: "migration-org", Name: "Scheduled", URL: "https://example.com", Protocol: "https", Status: models.MonitorStatusUp, IntervalSeconds: 10, TimeoutSeconds: 1, LastCheckAt: &past, CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateMonitor(context.Background(), &monitor); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := repo.DeleteMonitor(context.Background(), monitor.ID); err != nil {
			t.Error(err)
		}
	}()
	claimed, err := repo.ClaimDueMonitors(context.Background(), now, 16)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 || claimed[0].ID != monitor.ID {
		t.Fatalf("due monitor not claimed: %v", claimed)
	}
	claimed, err = repo.ClaimDueMonitors(context.Background(), now, 16)
	if err != nil || len(claimed) != 0 {
		t.Fatalf("active claim duplicated: %v", err)
	}
}
