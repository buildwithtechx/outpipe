package repositories

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
)

func TestMissingStatusPageDoesNotLogDatabaseError(t *testing.T) {
	connection, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	var output bytes.Buffer
	db, err := gorm.Open(sqlite.Dialector{Conn: connection}, &gorm.Config{
		Logger: logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Error}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.UptimeStatusPage{}); err != nil {
		t.Fatal(err)
	}
	repo, err := NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	page, err := repo.GetStatusPageBySlug(context.Background(), "default")
	if page != nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing page did not preserve not-found semantics: %v", err)
	}
	if output.Len() != 0 {
		t.Fatal("expected missing-page lookup logged a database error")
	}
}
