package integration_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/workers"
)

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.Session{}, &models.APIKey{}, &models.DeviceLogin{}); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestCleanupJobDeletesExpiredDatabaseRows(t *testing.T) {
	db := openTestDatabase(t)
	sessions, err := repositories.NewSessionRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	keys, err := repositories.NewAPIKeyRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	deviceLogins, err := repositories.NewDeviceLoginRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	expired := now.Add(-time.Hour)

	for _, session := range []models.Session{
		{Base: models.Base{ID: "session-expired"}, UserID: "user-1", TokenHash: "expired", ExpiresAt: now.Add(-time.Hour)},
		{Base: models.Base{ID: "session-revoked"}, UserID: "user-1", TokenHash: "revoked", ExpiresAt: now.Add(time.Hour), RevokedAt: &now},
		{Base: models.Base{ID: "session-active"}, UserID: "user-1", TokenHash: "active", ExpiresAt: now.Add(time.Hour)},
	} {

		if err := sessions.Create(context.Background(), &session); err != nil {
			t.Fatal(err)
		}
	}

	for _, key := range []models.APIKey{
		{Base: models.Base{ID: "key-expired"}, UserID: "user-1", Name: "expired", Prefix: "exp-prefix", SecretHash: "expired-hash", ExpiresAt: &expired},
		{Base: models.Base{ID: "key-revoked"}, UserID: "user-1", Name: "revoked", Prefix: "rev-prefix", SecretHash: "revoked-hash", RevokedAt: &now},
		{Base: models.Base{ID: "key-active"}, UserID: "user-1", Name: "active", Prefix: "act-prefix", SecretHash: "active-hash"},
	} {

		if err := keys.Create(context.Background(), &key); err != nil {
			t.Fatal(err)
		}
	}

	for _, login := range []models.DeviceLogin{
		{Base: models.Base{ID: "device-expired"}, CodeHash: "expired-code", ExpiresAt: now.Add(-time.Hour)},
		{Base: models.Base{ID: "device-active"}, CodeHash: "active-code", ExpiresAt: now.Add(time.Hour)},
	} {

		if err := deviceLogins.Create(context.Background(), &login); err != nil {
			t.Fatal(err)
		}
	}

	job, err := workers.NewCleanupJob(sessions, keys, deviceLogins)

	if err != nil {
		t.Fatal(err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("cleanup job: %v", err)
	}

	assertRowMissing(t, db, &models.Session{}, "session-expired")
	assertRowMissing(t, db, &models.Session{}, "session-revoked")
	assertRowPresent(t, db, &models.Session{}, "session-active")
	assertRowMissing(t, db, &models.APIKey{}, "key-expired")
	assertRowMissing(t, db, &models.APIKey{}, "key-revoked")
	assertRowPresent(t, db, &models.APIKey{}, "key-active")
	assertRowMissing(t, db, &models.DeviceLogin{}, "device-expired")
	assertRowPresent(t, db, &models.DeviceLogin{}, "device-active")

	if _, err := sessions.FindActive(context.Background(), "active", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("active session must remain usable: %v", err)
	}

	if _, err := deviceLogins.FindPending(context.Background(), "active-code", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("active device login must remain usable: %v", err)
	}
}

func assertRowPresent(t *testing.T, db *gorm.DB, model any, id string) {
	t.Helper()
	var count int64

	if err := db.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("expected row %q to be present, got %d rows", id, count)
	}
}

func assertRowMissing(t *testing.T, db *gorm.DB, model any, id string) {
	t.Helper()
	var count int64

	if err := db.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("expected row %q to be deleted, got %d rows", id, count)
	}
}
