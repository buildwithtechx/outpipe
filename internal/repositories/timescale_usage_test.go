package repositories

import (
	"context"
	"testing"
	"time"

	sqliteGorm "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
)

func suppressSQLiteUnused(err *sqlite.Error) {}

func TestGormTimeSeriesUsageRepository(t *testing.T) {
	db, err := gorm.Open(sqliteGorm.Open(":memory:"), &gorm.Config{})

	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.Exec(`
		CREATE TABLE usage_analytics_events (
			id TEXT PRIMARY KEY,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			organization_id TEXT NOT NULL,
			tunnel_id TEXT,
			event_type TEXT NOT NULL,
			bytes INTEGER NOT NULL DEFAULT 0,
			connections INTEGER NOT NULL DEFAULT 0,
			occurred_at DATETIME NOT NULL
		);
	`).Error

	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	repo, err := NewGormTimeSeriesUsageRepository(db)

	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	now := time.Now()
	event := models.UsageEvent{
		OrganizationID: "org-ts-1",
		EventType:      "http_request",
		Bytes:          2048,
		Connections:    1,
		OccurredAt:     now,
	}

	if err := repo.RecordTimeSeriesEvent(context.Background(), event); err != nil {
		t.Fatalf("RecordTimeSeriesEvent failed: %v", err)
	}

	snapshot, err := repo.AggregateTimeSeries(context.Background(), "org-ts-1", now.Add(-1*time.Hour), now.Add(1*time.Hour))

	if err != nil {
		t.Fatalf("AggregateTimeSeries failed: %v", err)
	}

	if snapshot.BandwidthBytes != 2048 {
		t.Fatalf("expected 2048 bandwidth bytes, got %d", snapshot.BandwidthBytes)
	}

	if snapshot.RequestCount != 1 {
		t.Fatalf("expected 1 request count, got %d", snapshot.RequestCount)
	}
}
