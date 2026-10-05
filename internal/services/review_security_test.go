package services

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func reviewTestDatabase(t *testing.T, tables ...any) *gorm.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestShareConcurrentRevealHonorsViewLimit(t *testing.T) {
	db := reviewTestDatabase(t, &models.SecretShareLink{})
	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewShareService(repo)
	if err != nil {
		t.Fatal(err)
	}
	link, err := service.CreateShare(context.Background(), CreateShareInput{Ciphertext: "encrypted", IV: "iv", KeyVerifier: "key", MaxViews: 1})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	var successes atomic.Int32
	start := make(chan struct{})
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			if _, err := service.RevealShare(context.Background(), link.ID, "key", nil); err == nil {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful reveals = %d", successes.Load())
	}
	stored, err := repo.FindShareLink(context.Background(), link.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Views != 1 {
		t.Fatalf("stored views = %d", stored.Views)
	}
}

func TestSecretSlugsScopesAndCorruption(t *testing.T) {
	db := reviewTestDatabase(t, &models.SecretProject{}, &models.SecretEnvironment{}, &models.SecretEntry{}, &models.SecretVersion{}, &models.SecretMachineToken{})
	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewSecretService(repo, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	project, err := service.CreateProject(ctx, "org", "backend", "Backend", "", "user")
	if err != nil {
		t.Fatal(err)
	}
	environment, err := service.CreateEnvironment(ctx, "org", project.Slug, "staging", "Staging", "user")
	if err != nil {
		t.Fatal(err)
	}
	secret, err := service.SetSecret(ctx, "org", project.Slug, environment.Slug, "API_KEY", "value", "", "user")
	if err != nil {
		t.Fatal(err)
	}
	items, err := service.ListSecrets(ctx, "org", project.Slug, environment.Slug, true)
	if err != nil || len(items) != 1 || items[0].Value != "value" {
		t.Fatalf("slug lookup failed: %v", err)
	}
	token, err := service.CreateMachineToken(ctx, "org", project.Slug, environment.Slug, "CI", []string{"secrets:read"}, "user")
	if err != nil {
		t.Fatal(err)
	}
	if token.Token.ProjectID == nil || *token.Token.ProjectID != project.ID || token.Token.EnvironmentID == nil || *token.Token.EnvironmentID != environment.ID {
		t.Fatal("token scope was not resolved")
	}
	if _, err := service.CreateMachineToken(ctx, "foreign", project.ID, environment.ID, "CI", nil, "user"); err == nil {
		t.Fatal("foreign project was accepted")
	}
	for _, values := range []map[string]any{{"ciphertext": "!"}, {"ciphertext": "dmFsdWU=", "iv": "!"}, {"ciphertext": "dmFsdWU=", "iv": "dmFsdWU="}} {
		if err := db.Model(&models.SecretVersion{}).Where("entry_id = ?", secret.ID).Updates(values).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := service.ListSecrets(ctx, "org", project.ID, environment.ID, true); err == nil {
			t.Fatal("corrupted secret returned success")
		}
	}
	if err := db.Where("entry_id = ?", secret.ID).Delete(&models.SecretVersion{}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.ListSecrets(ctx, "org", project.ID, environment.ID, true); err == nil {
		t.Fatal("missing secret version returned success")
	}
}
