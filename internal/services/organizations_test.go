package services

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestCreateOrganizationPersistsOwnerAndMembership(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:organization-create?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Organization{}, &models.OrganizationMember{}); err != nil {
		t.Fatal(err)
	}

	repo, err := repositories.NewOrganizationRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewOrganizationService(repo)
	if err != nil {
		t.Fatal(err)
	}

	ownerID := "1c98af02-620f-4565-9244-ab3ffdee36fe"
	organization, err := service.Create(context.Background(), ownerID, "Labs", "labs")
	if err != nil {
		t.Fatal(err)
	}
	if organization.OwnerID != ownerID {
		t.Fatalf("returned owner ID = %q, want %q", organization.OwnerID, ownerID)
	}

	stored, err := repo.FindByID(context.Background(), organization.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.OwnerID != ownerID {
		t.Fatalf("stored owner ID = %q, want %q", stored.OwnerID, ownerID)
	}

	member, err := repo.FindMember(context.Background(), organization.ID, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	if member.Role != models.MemberRoleOwner {
		t.Fatalf("owner member role = %q, want %q", member.Role, models.MemberRoleOwner)
	}
}
