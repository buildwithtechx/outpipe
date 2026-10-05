package services

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestSecretConcurrentWritesRestoreAndProjectDeletion(t *testing.T) {
	db := reviewTestDatabase(t, &models.SecretProject{}, &models.SecretEnvironment{}, &models.SecretEntry{}, &models.SecretVersion{}, &models.SecretMachineToken{}, &models.SecretShareLink{})
	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewSecretService(repo, " "); err == nil {
		t.Fatal("blank encryption key accepted")
	}
	s, err := NewSecretService(repo, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := s.CreateProject(ctx, "org", "$(command)", "bad", "", "user"); err == nil {
		t.Fatal("unsafe slug accepted")
	}
	project, err := s.CreateProject(ctx, "org", "backend", "Backend", "", "user")
	if err != nil {
		t.Fatal(err)
	}
	environment, err := s.CreateEnvironment(ctx, "org", project.ID, "production", "Production", "user")
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	start := make(chan struct{})
	for range 24 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			if _, err := s.SetSecret(ctx, "org", project.ID, environment.ID, "KEY", "  exact value  ", "", "user"); err != nil {
				t.Error(err)
			}
		}()
	}
	close(start)
	wait.Wait()
	items, err := s.ListSecrets(ctx, "org", project.ID, environment.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Version != 24 || items[0].Value != "  exact value  " {
		t.Fatalf("concurrent writes lost: %#v", items)
	}
	if err := s.SoftDeleteSecret(ctx, "org", items[0].ID); err != nil {
		t.Fatal(err)
	}
	trash, err := s.ListTrash(ctx, "org", "", "")
	if err != nil || len(trash) != 1 {
		t.Fatalf("organization trash: %v, %v", trash, err)
	}
	if _, err := s.SetSecret(ctx, "org", project.ID, environment.ID, "KEY", "replacement", "", "user"); err != nil {
		t.Fatal(err)
	}
	if err := s.RestoreSecret(ctx, "org", items[0].ID); err == nil {
		t.Fatal("restore duplicated an active key")
	}
	if err := s.DeleteProject(ctx, "org", project.ID); err != nil {
		t.Fatal(err)
	}
	first, err := s.CreateProject(ctx, "org", "duplicate", "First", "", "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateProject(ctx, "org", "duplicate", "Second", "", "user"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ResolveScope(ctx, "org", "duplicate", ""); err == nil {
		t.Fatal("ambiguous slug selected a project")
	}
	if resolved, _, err := s.ResolveScope(ctx, "org", first.ID, ""); err != nil || resolved != first.ID {
		t.Fatal("explicit project ID did not resolve")
	}
	for _, model := range []any{&models.SecretEntry{}, &models.SecretVersion{}, &models.SecretEnvironment{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("project data retained: %T", model)
		}
	}
}

func TestShareOrganizationListExposesOnlyMetadata(t *testing.T) {
	db := reviewTestDatabase(t, &models.SecretShareLink{})
	repo, err := repositories.NewSecretRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewShareService(repo)
	if err != nil {
		t.Fatal(err)
	}
	org := "org"
	link := models.SecretShareLink{ID: "share", OrganizationID: &org, Ciphertext: "sensitive", IV: "nonce", KeyVerifier: "verifier"}
	if err := repo.CreateShareLink(context.Background(), &link); err != nil {
		t.Fatal(err)
	}
	items, err := s.ListOrgShares(context.Background(), org)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"ciphertext", "keyVerifier", "passwordVerifier", "passwordSalt", "\"iv\""} {
		if strings.Contains(string(encoded), field) {
			t.Fatalf("share list exposed %s", field)
		}
	}
}
