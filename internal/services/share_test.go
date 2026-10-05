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

func TestShareServiceLifecycle(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.SecretShareLink{}); err != nil {
		t.Fatal(err)
	}

	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewShareService(repo)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	salt := "salt-123"
	pwdVerifier := "pwd-hash-abc"
	input := CreateShareInput{
		Ciphertext:       "ciphertext-xyz",
		IV:               "iv-123",
		KeyVerifier:      "verifier-456",
		PasswordSalt:     &salt,
		PasswordVerifier: &pwdVerifier,
		ContentFormat:    "bundle",
		TTLSeconds:       3600,
		MaxViews:         2,
	}

	link, err := service.CreateShare(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if link.ID == "" {
		t.Fatal("expected share link ID to be generated")
	}

	meta, err := service.GetShareMeta(ctx, link.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !meta.NeedsPassword {
		t.Fatal("expected needsPassword to be true")
	}
	if meta.ViewsRemaining != 2 {
		t.Fatalf("expected viewsRemaining 2, got %d", meta.ViewsRemaining)
	}

	// Reveal with wrong password
	wrongPwd := "wrong-pwd"
	_, err = service.RevealShare(ctx, link.ID, "verifier-456", &wrongPwd)
	if err == nil {
		t.Fatal("expected error with wrong password")
	}

	// Reveal 1 with correct password
	res1, err := service.RevealShare(ctx, link.ID, "verifier-456", &pwdVerifier)
	if err != nil {
		t.Fatalf("reveal failed: %v", err)
	}
	if res1.Ciphertext != "ciphertext-xyz" {
		t.Fatalf("expected ciphertext-xyz, got %s", res1.Ciphertext)
	}

	// Reveal 2 (max view reached)
	res2, err := service.RevealShare(ctx, link.ID, "verifier-456", &pwdVerifier)
	if err != nil {
		t.Fatalf("reveal 2 failed: %v", err)
	}
	if res2.Ciphertext != "ciphertext-xyz" {
		t.Fatalf("expected ciphertext-xyz, got %s", res2.Ciphertext)
	}

	// Reveal 3 should fail as views are exhausted
	_, err = service.RevealShare(ctx, link.ID, "verifier-456", &pwdVerifier)
	if err == nil {
		t.Fatal("expected error when views exhausted")
	}
}
