package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func setupQuotaDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	if err := db.AutoMigrate(&models.OrganizationMonthlyUsage{}, &models.OrganizationTelemetryReceipt{}); err != nil {
		t.Fatalf("migrate quota tables: %v", err)
	}
	return db
}

func TestAdmitBatchDeduplicatesIdenticalRetries(t *testing.T) {
	repo, err := repositories.NewTelemetryQuotaRepository(setupQuotaDB(t))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	duplicate, err := repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-1", 100, 1000)
	if err != nil || duplicate {
		t.Fatalf("first admission duplicate=%v: %v", duplicate, err)
	}
	duplicate, err = repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-1", 100, 1000)
	if err != nil || !duplicate {
		t.Fatalf("retry admission duplicate=%v: %v", duplicate, err)
	}
	usage, err := repo.UsageForMonth(ctx, "org-1", "2026-10-01")
	if err != nil {
		t.Fatal(err)
	}
	if usage.Used != 100 {
		t.Fatalf("expected single charge of 100 bytes, got %d", usage.Used)
	}
}

func TestAdmitBatchEnforcesMonthlyLimit(t *testing.T) {
	repo, err := repositories.NewTelemetryQuotaRepository(setupQuotaDB(t))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-1", 900, 1000); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-2", 200, 1000); !errors.Is(err, repositories.ErrTelemetryQuotaExceeded) {
		t.Fatalf("expected quota exceeded, got %v", err)
	}
}

func TestRefundBatchRestoresQuotaAfterFailure(t *testing.T) {
	repo, err := repositories.NewTelemetryQuotaRepository(setupQuotaDB(t))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-1", 400, 1000); err != nil {
		t.Fatal(err)
	}
	if err := repo.RefundBatch(ctx, "org-1", "2026-10-01", "key-1", 400); err != nil {
		t.Fatal(err)
	}
	usage, err := repo.UsageForMonth(ctx, "org-1", "2026-10-01")
	if err != nil {
		t.Fatal(err)
	}
	if usage.Used != 0 {
		t.Fatalf("expected refunded usage 0, got %d", usage.Used)
	}
}

func TestDeleteExpiredReceiptsKeepsMonthlyCounters(t *testing.T) {
	db := setupQuotaDB(t)
	repo, err := repositories.NewTelemetryQuotaRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := repo.AdmitBatch(ctx, "org-1", "2026-10-01", "key-old", 50, 1000); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -121)
	if err := db.Model(&models.OrganizationTelemetryReceipt{}).Where("organization_id = ? AND receipt_key = ?", "org-1", "key-old").Update("created_at", old).Error; err != nil {
		t.Fatal(err)
	}
	removed, err := repo.DeleteExpiredReceipts(ctx, time.Now().UTC().AddDate(0, 0, -120))
	if err != nil || removed != 1 {
		t.Fatalf("expected 1 expired receipt removed, got %d: %v", removed, err)
	}
	if _, err := repo.UsageForMonth(ctx, "org-1", "2026-10-01"); err != nil {
		t.Fatalf("monthly counter must survive receipt expiry: %v", err)
	}
}
