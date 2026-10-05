package handlers

import (
	"database/sql"
	"encoding/json"
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

func TestPublicStatusDistinguishesMissingPageAndDatabaseFailure(t *testing.T) {
	for _, available := range []bool{false, true} {
		t.Run(map[bool]string{false: "database failure", true: "missing page"}[available], func(t *testing.T) {
			connection, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := connection.Close(); err != nil {
					t.Errorf("close database: %v", err)
				}
			})
			db, err := gorm.Open(sqlite.Dialector{Conn: connection}, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			if available {
				if err := db.AutoMigrate(&models.UptimeStatusPage{}); err != nil {
					t.Fatal(err)
				}
			}
			repo, err := repositories.NewUptimeRepository(db)
			if err != nil {
				t.Fatal(err)
			}
			service, err := services.NewUptimeService(repo)
			if err != nil {
				t.Fatal(err)
			}
			handler, err := NewUptimeHandler(service)
			if err != nil {
				t.Fatal(err)
			}
			app := fiber.New()
			app.Get("/status/:slug", handler.GetPublicStatus)
			response, err := app.Test(httptest.NewRequest("GET", "/status/default", nil))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := response.Body.Close(); err != nil {
					t.Errorf("close response: %v", err)
				}
			})
			want := map[bool]int{false: 500, true: 404}[available]
			if response.StatusCode != want {
				t.Fatalf("expected %d, got %d", want, response.StatusCode)
			}
			var result ErrorResponse
			if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
			if !available && result.Error != "Status is temporarily unavailable" {
				t.Fatal("public response exposed database error details")
			}
		})
	}
}
