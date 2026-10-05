package services

import (
	"context"
	"fmt"
	"strings"

	"outpipe.dev/outpipe/internal/repositories"
)

func (s *SecretService) ResolveScope(ctx context.Context, orgID, project, environment string) (string, string, error) {
	return resolveSecretScope(ctx, s.repo, orgID, project, environment)
}

func resolveSecretScope(ctx context.Context, repo repositories.SecretRepository, orgID, project, environment string) (string, string, error) {
	project, environment = strings.TrimSpace(project), strings.TrimSpace(environment)
	if project == "" {
		if environment != "" {
			return "", "", fmt.Errorf("environment requires a project")
		}
		return "", "", nil
	}
	projects, err := repo.ListProjects(ctx, orgID)
	if err != nil {
		return "", "", fmt.Errorf("resolve secret project: %w", err)
	}
	projectID := ""
	for _, item := range projects {
		if item.ID == project || item.Slug == project {
			projectID = item.ID
			break
		}
	}
	if projectID == "" {
		return "", "", fmt.Errorf("secret project: %w", repositories.ErrNotFound)
	}
	if environment == "" {
		return projectID, "", nil
	}
	environments, err := repo.ListEnvironments(ctx, orgID, projectID)
	if err != nil {
		return "", "", fmt.Errorf("resolve secret environment: %w", err)
	}
	for _, item := range environments {
		if item.ID == environment || item.Slug == environment {
			return projectID, item.ID, nil
		}
	}
	return "", "", fmt.Errorf("secret environment: %w", repositories.ErrNotFound)
}
