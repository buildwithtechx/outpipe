package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestUsageSnapshotReturnsEmptyPeriodWhenNoSnapshotExists(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:usage-snapshot?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.UsageSnapshot{}, &models.UsageEvent{}); err != nil {
		t.Fatal(err)
	}
	repo, err := repositories.NewUsageRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewUsageService(repo)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewUsageHandler(service)
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Get("/organizations/:organizationID/usage/snapshot", handler.Snapshot)
	periodStart := "2026-09-23T09:00:00Z"
	request := httptest.NewRequest(http.MethodGet, "/organizations/org-1/usage/snapshot?periodStart="+periodStart, nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	var snapshot models.UsageSnapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.OrganizationID != "org-1" || !snapshot.PeriodStart.Equal(time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC)) || !snapshot.PeriodEnd.Equal(time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected empty snapshot period: %+v", snapshot)
	}
	if snapshot.RequestCount != 0 || snapshot.ErrorCount != 0 || snapshot.ActiveConnections != 0 || snapshot.BandwidthBytes != 0 {
		t.Fatalf("empty snapshot has nonzero usage: %+v", snapshot)
	}

	tunnelID := "7fe53563-c9c8-4fe2-aacd-615812f521e1"
	event := models.UsageEvent{
		OrganizationID: "org-1",
		TunnelID:       &tunnelID,
		EventType:      "request",
		Bytes:          120,
		Connections:    2,
		StatusCode:     200,
		OccurredAt:     time.Date(2026, 9, 23, 9, 30, 0, 0, time.UTC),
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	errorEvent := models.UsageEvent{
		OrganizationID: "org-1",
		TunnelID:       &tunnelID,
		EventType:      "request",
		Bytes:          80,
		Connections:    3,
		StatusCode:     404,
		OccurredAt:     time.Date(2026, 9, 23, 9, 45, 0, 0, time.UTC),
	}
	if err := db.Create(&errorEvent).Error; err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest(http.MethodGet, "/organizations/org-1/usage/snapshot?periodStart="+periodStart, nil)
	response, err = app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status with event = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.OrganizationID != "org-1" || !snapshot.PeriodStart.Equal(time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC)) || !snapshot.PeriodEnd.Equal(time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected computed snapshot period: %+v", snapshot)
	}
	if snapshot.RequestCount != 2 || snapshot.ErrorCount != 1 || snapshot.TunnelCount != 1 || snapshot.ActiveConnections != 5 || snapshot.BandwidthBytes != 200 {
		t.Fatalf("computed snapshot = %+v, want two requests, one error, one tunnel, five connections, and 200 bytes", snapshot)
	}
}
