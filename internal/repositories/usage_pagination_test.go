package repositories

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func TestListEventsUsesStableCursorOrderingAndLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:usage-pagination?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.UsageEvent{}); err != nil {
		t.Fatal(err)
	}

	occurredAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	for _, id := range []string{"event-1", "event-2", "event-3"} {
		if err := db.Create(&models.UsageEvent{Base: models.Base{ID: id}, OrganizationID: "org-1", EventType: "request", OccurredAt: occurredAt}).Error; err != nil {
			t.Fatal(err)
		}
	}

	repository := &GormUsageRepository{db: db}
	first, err := repository.ListEvents(context.Background(), "org-1", occurredAt.Add(-time.Minute), occurredAt.Add(time.Minute), 2, time.Time{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 2 || first[0].ID != "event-1" || first[1].ID != "event-2" {
		t.Fatalf("first page = %#v, want event-1 and event-2", first)
	}

	second, err := repository.ListEvents(context.Background(), "org-1", occurredAt.Add(-time.Minute), occurredAt.Add(time.Minute), 2, first[1].OccurredAt, first[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 1 || second[0].ID != "event-3" {
		t.Fatalf("second page = %#v, want event-3", second)
	}
}
