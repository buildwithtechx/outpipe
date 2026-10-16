package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AIAgentService struct {
	repo repositories.AIAgentRepository
	obs  *ObservabilityService
	cfg  config.AgentConfig
}

func NewAIAgentService(repo repositories.AIAgentRepository, obs *ObservabilityService, cfgs ...config.AgentConfig) (*AIAgentService, error) {
	if repo == nil {
		return nil, fmt.Errorf("ai agent repository is required")
	}
	svc := &AIAgentService{repo: repo, obs: obs}
	if len(cfgs) > 0 {
		svc.cfg = cfgs[0]
	}
	return svc, nil
}

func (s *AIAgentService) ListThreads(ctx context.Context, orgID, userID string) ([]models.AIAgentThread, error) {
	threads, err := s.repo.ListThreads(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("list threads: %w", err)
	}
	return threads, nil
}

func (s *AIAgentService) GetThread(ctx context.Context, orgID, threadID string) (*models.AIAgentThread, error) {
	thread, err := s.repo.GetThread(ctx, orgID, threadID)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}
	return thread, nil
}

func (s *AIAgentService) DeleteThread(ctx context.Context, orgID, threadID, userID string) error {
	if err := s.repo.DeleteThread(ctx, orgID, threadID, userID); err != nil {
		return fmt.Errorf("delete thread: %w", err)
	}
	return nil
}

func (s *AIAgentService) providerKey(provider string) string {
	switch provider {
	case "openai":
		return s.firstNonEmpty(s.cfg.OpenAIKey, os.Getenv("OPENAI_API_KEY"))
	case "xai":
		return s.firstNonEmpty(s.cfg.XAIKey, os.Getenv("XAI_API_KEY"))
	case "anthropic":
		return s.firstNonEmpty(s.cfg.AnthropicKey, os.Getenv("ANTHROPIC_API_KEY"))
	default:
		return ""
	}
}

func (s *AIAgentService) firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (s *AIAgentService) GetConfig() models.AIAgentConfig {
	openAIKey, xaiKey, anthropicKey := s.providerKey("openai"), s.providerKey("xai"), s.providerKey("anthropic")
	available := []string{"mock"}
	if openAIKey != "" {
		available = append(available, "openai")
	}
	if xaiKey != "" {
		available = append(available, "xai")
	}
	if anthropicKey != "" {
		available = append(available, "anthropic")
	}

	provider := strings.ToLower(s.firstNonEmpty(s.cfg.DefaultProvider, os.Getenv("DEFAULT_AGENT_PROVIDER"), os.Getenv("AGENT_PROVIDER")))
	if provider == "" {
		if openAIKey != "" {
			provider = "openai"
		} else if xaiKey != "" {
			provider = "xai"
		} else if anthropicKey != "" {
			provider = "anthropic"
		} else {
			provider = "mock"
		}
	}
	model := s.firstNonEmpty(s.cfg.Model, os.Getenv("AGENT_MODEL"))
	if model == "" {
		switch provider {
		case "openai":
			model = "gpt-4o"
		case "xai":
			model = "grok-4.7"
		case "anthropic":
			model = "claude-3-5-sonnet-20241022"
		default:
			model = "outpipe-heuristic-engine-v1"
		}
	}
	return models.AIAgentConfig{
		Provider:           provider,
		Model:              model,
		Enabled:            true,
		AvailableProviders: available,
	}
}

func (s *AIAgentService) InvestigateStream(ctx context.Context, orgID, userID string, req models.AIChatRequest, emit func(models.AIStreamEvent) error) error {
	threadID := req.ThreadID
	var existingThread *models.AIAgentThread
	if threadID == "" {
		threadID = uuid.NewString()
		title := req.Message
		if len(title) > 40 {
			title = title[:40] + "…"
		}
		newThread := &models.AIAgentThread{
			Base:             models.Base{ID: threadID},
			OrganizationID:   orgID,
			UserID:           userID,
			Title:            title,
			SourceResourceID: req.SourceResourceID,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := s.repo.CreateThread(ctx, newThread); err != nil {
			return fmt.Errorf("create initial thread: %w", err)
		}
		existingThread = newThread
	} else {
		loaded, err := s.repo.GetThread(ctx, orgID, threadID)
		if err != nil {
			return fmt.Errorf("load thread %s: %w", threadID, err)
		}
		if loaded.OrganizationID != orgID || (userID != "" && loaded.UserID != "" && loaded.UserID != userID) {
			return fmt.Errorf("unauthorized access to thread %s", threadID)
		}
		existingThread = loaded
	}

	userMsg := &models.AIAgentMessage{
		Base:      models.Base{ID: uuid.NewString()},
		ThreadID:  threadID,
		Role:      "user",
		Content:   req.Message,
		Status:    "complete",
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateMessage(ctx, userMsg); err != nil {
		return fmt.Errorf("save user message: %w", err)
	}

	runID := uuid.NewString()
	_ = emit(models.AIStreamEvent{Type: "run", RunID: runID})

	step1 := &models.AIAgentStep{ID: "step-1", Label: "Reading telemetry & error status", Status: "running"}
	_ = emit(models.AIStreamEvent{Type: "step", Step: step1})

	var evidenceList []models.AIAgentEvidence
	if req.SourceResourceID != "" && s.obs != nil {
		if cap, err := s.obs.GetCapture(ctx, orgID, req.SourceResourceID); err == nil && cap != nil {
			evidenceList = append(evidenceList, models.AIAgentEvidence{
				ID:         "ev-req",
				Label:      fmt.Sprintf("%s %s (HTTP %d)", cap.Method, cap.Path, cap.StatusCode),
				Href:       fmt.Sprintf("/observability?id=%s", cap.ID),
				ObservedAt: cap.Timestamp.Format(time.RFC3339),
				Presentation: map[string]interface{}{
					"kind":         "request",
					"method":       cap.Method,
					"path":         cap.Path,
					"statusCode":   cap.StatusCode,
					"durationMs":   cap.DurationMs,
					"captureState": "full",
				},
			})
		}
	}

	step1.Status = "complete"
	step1.Detail = "Telemetry context verified"
	_ = emit(models.AIStreamEvent{Type: "step", Step: step1})

	if len(evidenceList) > 0 {
		_ = emit(models.AIStreamEvent{Type: "evidence", Evidence: evidenceList})
	}

	cfg := s.GetConfig()
	provider := cfg.Provider
	if req.Provider != "" {
		provider = req.Provider
	}

	systemPrompt := "You are Outpipe Agent, an expert read-only observability and tunneling investigator. Provide structured Markdown conclusions with ### Diagnostic Summary, ### Key Observations, ### Unconfirmed Possibilities, ### Recommended Next Checks."
	if len(evidenceList) > 0 {
		evidenceBytes, _ := json.Marshal(evidenceList)
		systemPrompt += "\n\nCollected Telemetry Evidence:\n" + string(evidenceBytes)
	}

	providerMessages := make([]map[string]string, 0, len(existingThread.Messages)+2)
	providerMessages = append(providerMessages, map[string]string{"role": "system", "content": systemPrompt})
	for _, m := range existingThread.Messages {
		if m.Content != "" {
			providerMessages = append(providerMessages, map[string]string{"role": m.Role, "content": m.Content})
		}
	}
	providerMessages = append(providerMessages, map[string]string{"role": "user", "content": req.Message})

	var fullResponse string
	var execErr error

	switch provider {
	case "openai":
		if key := s.providerKey("openai"); key != "" {
			fullResponse, execErr = s.streamOpenAI(ctx, key, cfg.Model, providerMessages, emit)
		} else {
			fullResponse, execErr = s.streamMock(ctx, req, emit)
		}
	case "xai":
		if key := s.providerKey("xai"); key != "" {
			fullResponse, execErr = s.streamOpenAICompat(ctx, "https://api.x.ai/v1/chat/completions", key, cfg.Model, providerMessages, emit)
		} else {
			fullResponse, execErr = s.streamMock(ctx, req, emit)
		}
	case "anthropic":
		if key := s.providerKey("anthropic"); key != "" {
			anthropicMessages := make([]map[string]string, 0, len(providerMessages))
			for _, pm := range providerMessages {
				if pm["role"] != "system" {
					anthropicMessages = append(anthropicMessages, pm)
				}
			}
			fullResponse, execErr = s.streamAnthropic(ctx, key, cfg.Model, systemPrompt, anthropicMessages, emit)
		} else {
			fullResponse, execErr = s.streamMock(ctx, req, emit)
		}
	default:
		fullResponse, execErr = s.streamMock(ctx, req, emit)
	}

	if execErr != nil {
		_ = emit(models.AIStreamEvent{Type: "finish", Status: "failed", Error: execErr.Error()})
		return execErr
	}

	evidenceBytes, _ := json.Marshal(evidenceList)
	assistantMsg := &models.AIAgentMessage{
		Base:         models.Base{ID: uuid.NewString()},
		ThreadID:     threadID,
		Role:         "assistant",
		Content:      fullResponse,
		Status:       "complete",
		EvidenceJSON: string(evidenceBytes),
		CreatedAt:    time.Now(),
	}
	_ = s.repo.CreateMessage(ctx, assistantMsg)

	_ = emit(models.AIStreamEvent{Type: "finish", Status: "complete"})
	return nil
}
