package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/infra/postgres"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func runBootstrapAdmin(ctx context.Context, cfg config.APIConfig, args []string) error {
	flags := flag.NewFlagSet("bootstrap-admin", flag.ContinueOnError)
	email := flags.String("email", "", "email address of the existing OAuth user")
	name := flags.String("name", "", "optional platform admin display name")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*email) == "" {
		return fmt.Errorf("--email is required")
	}
	db, err := postgres.Open(ctx, postgres.Config{DSN: cfg.Database.URL, MaxOpenConns: cfg.Database.MaxConns, MaxIdleConns: cfg.Database.MaxConns, ConnMaxLifetime: cfg.Database.MaxLifetime})
	if err != nil {
		return err
	}
	if err := postgres.Migrate(db); err != nil {
		return err
	}
	users, err := repositories.NewUserRepository(db)
	if err != nil {
		return err
	}
	admins, err := repositories.NewAdminRepository(db)
	if err != nil {
		return err
	}
	user, err := users.FindByEmail(ctx, *email)
	if err != nil {
		if err == repositories.ErrNotFound {
			return fmt.Errorf("user must sign in with OAuth before bootstrapping admin access")
		}
		return fmt.Errorf("find admin user: %w", err)
	}
	isAdmin, err := admins.IsPlatformAdmin(ctx, user.ID)
	if err != nil {
		return err
	}
	if isAdmin {
		return fmt.Errorf("user is already a platform administrator")
	}
	displayName := strings.TrimSpace(*name)
	if displayName == "" {
		displayName = user.Name
	}
	admin := &models.PlatformAdmin{UserID: user.ID, Name: displayName, Role: models.PlatformAdminOwner, Active: true}
	if err := admins.CreatePlatformAdmin(ctx, admin); err != nil {
		return err
	}
	fmt.Printf("platform administrator provisioned for %s\n", user.Email)
	return nil
}
