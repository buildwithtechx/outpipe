package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func setupTestAIHandler(t *testing.T) (*AIAgentHandler, *services.AIAgentService, func()) {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: conn}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	repo, err := repositories.NewAIAgentRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	svc, err := services.NewAIAgentService(repo, nil)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	h, err := NewAIAgentHandler(svc)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	return h, svc, func() { _ = conn.Close() }
}

func TestAIAgentHandlerListThreads(t *testing.T) {
	h, _, cleanup := setupTestAIHandler(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/organizations/:organizationID/agent/threads", h.ListThreads)

	req := httptest.NewRequest("GET", "/organizations/org-test/agent/threads", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["configured"] != true {
		t.Errorf("expected configured: true, got %v", body["configured"])
	}
}

func TestAIAgentHandlerChatStream(t *testing.T) {
	h, _, cleanup := setupTestAIHandler(t)
	defer cleanup()

	app := fiber.New()
	app.Post("/organizations/:organizationID/agent/chat", h.ChatStream)

	chatReq := models.AIChatRequest{
		Message: "Explain why status code is 504",
	}
	bodyBytes, _ := json.Marshal(chatReq)

	req := httptest.NewRequest("POST", "/organizations/org-test/agent/chat", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("chat request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if !bytes.Contains(respBody, []byte("data: ")) {
		t.Errorf("expected sse stream with 'data: ', got: %s", string(respBody))
	}
}
