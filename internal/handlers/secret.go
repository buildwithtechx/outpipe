package handlers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/services"
	"outpipe.dev/outpipe/internal/validation"
)

type SecretHandler struct {
	secrets *services.SecretService
}

func NewSecretHandler(secrets *services.SecretService) (*SecretHandler, error) {
	if secrets == nil {
		return nil, fmt.Errorf("secret service is required")
	}
	return &SecretHandler{secrets: secrets}, nil
}

type CreateProjectRequest struct {
	Slug        string `json:"slug" validate:"required,min=2,max=120"`
	Name        string `json:"name" validate:"required,max=120"`
	Description string `json:"description"`
}

type CreateEnvironmentRequest struct {
	Slug string `json:"slug" validate:"required,min=2,max=64"`
	Name string `json:"name" validate:"required,max=64"`
}

type SetSecretRequest struct {
	Key     string `json:"key" validate:"required,max=256"`
	Value   string `json:"value" validate:"required"`
	Comment string `json:"comment"`
}

type CreateMachineTokenRequest struct {
	ProjectID     string   `json:"projectId"`
	EnvironmentID string   `json:"environmentId"`
	Name          string   `json:"name" validate:"required,max=120"`
	Scopes        []string `json:"scopes"`
}

func (h *SecretHandler) ListProjects(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projects, err := h.secrets.ListProjects(c.UserContext(), orgID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.JSON(projects)
}

func (h *SecretHandler) CreateProject(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	var input CreateProjectRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode project request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	project, err := h.secrets.CreateProject(c.UserContext(), orgID, input.Slug, input.Name, input.Description, userID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *SecretHandler) DeleteProject(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	if err := h.secrets.DeleteProject(c.UserContext(), orgID, projectID); err != nil {
		return writeSecretError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SecretHandler) ListEnvironments(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	envs, err := h.secrets.ListEnvironments(c.UserContext(), orgID, projectID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.JSON(envs)
}

func (h *SecretHandler) CreateEnvironment(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	var input CreateEnvironmentRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode environment request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	env, err := h.secrets.CreateEnvironment(c.UserContext(), orgID, projectID, input.Slug, input.Name, userID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(env)
}

func (h *SecretHandler) ListSecrets(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	envID := strings.TrimSpace(c.Params("environmentID"))
	reveal := c.Query("reveal") == "true"
	secrets, err := h.secrets.ListSecrets(c.UserContext(), orgID, projectID, envID, reveal)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.JSON(secrets)
}

func (h *SecretHandler) SetSecret(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	envID := strings.TrimSpace(c.Params("environmentID"))
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	var input SetSecretRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode secret request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	item, err := h.secrets.SetSecret(c.UserContext(), orgID, projectID, envID, input.Key, input.Value, input.Comment, userID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *SecretHandler) DeleteSecret(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	entryID := strings.TrimSpace(c.Params("secretID"))
	if err := h.secrets.SoftDeleteSecret(c.UserContext(), orgID, entryID); err != nil {
		return writeSecretError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SecretHandler) RestoreSecret(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	entryID := strings.TrimSpace(c.Params("secretID"))
	if err := h.secrets.RestoreSecret(c.UserContext(), orgID, entryID); err != nil {
		return writeSecretError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *SecretHandler) ListTrash(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	projectID := strings.TrimSpace(c.Params("projectID"))
	envID := strings.TrimSpace(c.Params("environmentID"))
	entries, err := h.secrets.ListTrash(c.UserContext(), orgID, projectID, envID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.JSON(entries)
}

func (h *SecretHandler) ListMachineTokens(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	tokens, err := h.secrets.ListMachineTokens(c.UserContext(), orgID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.JSON(tokens)
}

func (h *SecretHandler) CreateMachineToken(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	userID, err := sessionUserID(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, err)
	}
	var input CreateMachineTokenRequest
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, fmt.Errorf("decode token request: %w", err))
	}
	if err := validation.Struct(input); err != nil {
		return writeError(c, fiber.StatusBadRequest, err)
	}
	token, err := h.secrets.CreateMachineToken(c.UserContext(), orgID, input.ProjectID, input.EnvironmentID, input.Name, input.Scopes, userID)
	if err != nil {
		return writeSecretError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(token)
}

func (h *SecretHandler) RevokeMachineToken(c *fiber.Ctx) error {
	orgID := strings.TrimSpace(c.Params("organizationID"))
	tokenID := strings.TrimSpace(c.Params("tokenID"))
	if err := h.secrets.RevokeMachineToken(c.UserContext(), orgID, tokenID); err != nil {
		return writeSecretError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SecretHandler) CLIGetSecrets(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	rawToken = strings.TrimSpace(rawToken)

	token, err := h.secrets.VerifyMachineToken(c.UserContext(), rawToken)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, fmt.Errorf("invalid token"))
	}
	var scopes []string
	if err := json.Unmarshal([]byte(token.Scopes), &scopes); err != nil || !slices.Contains(scopes, "secrets:read") {
		return writeError(c, fiber.StatusForbidden, fmt.Errorf("secrets:read scope is required"))
	}

	projectID := c.Query("project")
	if token.ProjectID != nil && *token.ProjectID != "" {
		projectID = *token.ProjectID
	}
	envID := c.Query("environment")
	if token.EnvironmentID != nil && *token.EnvironmentID != "" {
		envID = *token.EnvironmentID
	}

	secrets, err := h.secrets.ListSecrets(c.UserContext(), token.OrganizationID, projectID, envID, true)
	if err != nil {
		return writeSecretError(c, err)
	}

	envMap := make(map[string]string, len(secrets))
	for _, s := range secrets {
		if _, exists := envMap[s.Key]; exists {
			return writeError(c, fiber.StatusBadRequest, fmt.Errorf("secret keys are ambiguous; specify project and environment"))
		}
		envMap[s.Key] = s.Value
	}
	return c.JSON(envMap)
}
