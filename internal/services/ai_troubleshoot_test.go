package services_test

import (
	"context"
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func setupTestAIAgentDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	repo, err := repositories.NewAIAgentRepository(db)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	_ = repo

	return db, func() { _ = sqlDB.Close() }
}

func TestAIAgentServiceGetConfig(t *testing.T) {
	db, cleanup := setupTestAIAgentDB(t)
	defer cleanup()

	repo, err := repositories.NewAIAgentRepository(db)
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	svc, err := services.NewAIAgentService(repo, nil)
	if err != nil {
		t.Fatalf("create svc: %v", err)
	}

	cfg := svc.GetConfig()
	if !cfg.Enabled {
		t.Errorf("expected agent service to be enabled")
	}
	if cfg.Provider == "" {
		t.Errorf("expected default provider to be set")
	}
	if len(cfg.AvailableProviders) == 0 {
		t.Errorf("expected at least mock provider in available providers")
	}
}

func TestAIAgentServiceInvestigateStream(t *testing.T) {
	db, cleanup := setupTestAIAgentDB(t)
	defer cleanup()

	repo, err := repositories.NewAIAgentRepository(db)
	if err != nil {
		t.Fatalf("create repo: %v", err)
	}

	svc, err := services.NewAIAgentService(repo, nil)
	if err != nil {
		t.Fatalf("create svc: %v", err)
	}

	ctx := context.Background()
	orgID := "org-test-1"
	userID := "user-test-1"

	var receivedEvents []models.AIStreamEvent
	emit := func(evt models.AIStreamEvent) error {
		receivedEvents = append(receivedEvents, evt)
		return nil
	}

	req := models.AIChatRequest{
		Message: "Why did request req-123 fail with 502 Bad Gateway?",
	}

	err = svc.InvestigateStream(ctx, orgID, userID, req, emit)
	if err != nil {
		t.Fatalf("InvestigateStream failed: %v", err)
	}

	if len(receivedEvents) == 0 {
		t.Fatalf("expected to receive stream events, got none")
	}

	hasRun := false
	hasStep := false
	hasText := false
	hasFinish := false

	for _, evt := range receivedEvents {
		switch evt.Type {
		case "run":
			hasRun = true
		case "step":
			hasStep = true
		case "text":
			hasText = true
		case "finish":
			hasFinish = true
		}
	}

	if !hasRun || !hasStep || !hasText || !hasFinish {
		t.Errorf("missing events: run=%v, step=%v, text=%v, finish=%v", hasRun, hasStep, hasText, hasFinish)
	}

	threads, err := svc.ListThreads(ctx, orgID, userID)
	if err != nil {
		t.Fatalf("list threads failed: %v", err)
	}

	if len(threads) == 0 {
		t.Fatalf("expected thread to be persisted in database")
	}

	threadID := threads[0].ID
	if err := svc.DeleteThread(ctx, orgID, threadID); err != nil {
		t.Fatalf("delete thread failed: %v", err)
	}

	threadsAfter, _ := svc.ListThreads(ctx, orgID, userID)
	if len(threadsAfter) != 0 {
		t.Errorf("expected 0 threads after delete, got %d", len(threadsAfter))
	}
}
