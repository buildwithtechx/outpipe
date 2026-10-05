package services

import (
	"context"
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestSecretServiceLifecycle(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.SecretProject{},
		&models.SecretEnvironment{},
		&models.SecretEntry{},
		&models.SecretVersion{},
		&models.SecretMachineToken{},
		&models.SecretAuditEvent{},
		&models.SecretShareLink{},
	); err != nil {
		t.Fatal(err)
	}

	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	encryptionKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	service, err := NewSecretService(repo, encryptionKey)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	orgID := "org-1"
	userID := "user-1"

	project, err := service.CreateProject(ctx, orgID, "backend-api", "Backend API", "Primary backend", userID)
	if err != nil {
		t.Fatal(err)
	}
	if project.Slug != "backend-api" {
		t.Fatalf("expected slug backend-api, got %s", project.Slug)
	}

	env, err := service.CreateEnvironment(ctx, orgID, project.ID, "production", "Production", userID)
	if err != nil {
		t.Fatal(err)
	}
	if env.Slug != "production" {
		t.Fatalf("expected slug production, got %s", env.Slug)
	}

	item, err := service.SetSecret(ctx, orgID, project.ID, env.ID, "DATABASE_URL", "postgres://user:pass@localhost:5432/db", "Main DB", userID)
	if err != nil {
		t.Fatal(err)
	}
	if item.Key != "DATABASE_URL" {
		t.Fatalf("expected key DATABASE_URL, got %s", item.Key)
	}
	if item.Value != "postgres://user:pass@localhost:5432/db" {
		t.Fatalf("expected decrypted value, got %s", item.Value)
	}

	secrets, err := service.ListSecrets(ctx, orgID, project.ID, env.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
	if secrets[0].Value != "postgres://user:pass@localhost:5432/db" {
		t.Fatalf("expected decrypted value in list, got %s", secrets[0].Value)
	}

	masked, err := service.ListSecrets(ctx, orgID, project.ID, env.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if masked[0].Value != "" {
		t.Fatalf("expected empty value when reveal=false, got %s", masked[0].Value)
	}

	createdToken, err := service.CreateMachineToken(ctx, orgID, project.ID, env.ID, "ci-runner", []string{"secrets:read"}, userID)
	if err != nil {
		t.Fatal(err)
	}
	if createdToken.Raw == "" {
		t.Fatal("expected raw token to be non-empty")
	}

	verified, err := service.VerifyMachineToken(ctx, createdToken.Raw)
	if err != nil {
		t.Fatalf("verify machine token failed: %v", err)
	}
	if verified.Name != "ci-runner" {
		t.Fatalf("expected token name ci-runner, got %s", verified.Name)
	}

	if err := service.SoftDeleteSecret(ctx, orgID, item.ID); err != nil {
		t.Fatal(err)
	}

	activeSecrets, err := service.ListSecrets(ctx, orgID, project.ID, env.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(activeSecrets) != 0 {
		t.Fatalf("expected 0 active secrets after delete, got %d", len(activeSecrets))
	}

	trash, err := service.ListTrash(ctx, orgID, project.ID, env.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(trash) != 1 {
		t.Fatalf("expected 1 trash item, got %d", len(trash))
	}

	if err := service.RestoreSecret(ctx, orgID, item.ID); err != nil {
		t.Fatal(err)
	}
	restored, err := service.ListSecrets(ctx, orgID, project.ID, env.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) != 1 {
		t.Fatalf("expected 1 restored secret, got %d", len(restored))
	}
}
