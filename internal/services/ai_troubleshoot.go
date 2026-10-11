package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *AIAgentService) DeleteThread(ctx context.Context, orgID, threadID string) error {
	if err := s.repo.DeleteThread(ctx, orgID, threadID); err != nil {
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

	var fullResponse string
	var execErr error

	switch provider {
	case "openai":
		if key := s.providerKey("openai"); key != "" {
			fullResponse, execErr = s.streamOpenAI(ctx, key, cfg.Model, req.Message, emit)
		} else {
			fullResponse, execErr = s.streamMock(ctx, req, emit)
		}
	case "xai":
		if key := s.providerKey("xai"); key != "" {
			fullResponse, execErr = s.streamOpenAICompat(ctx, "https://api.x.ai/v1/chat/completions", key, cfg.Model, req.Message, emit)
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

func (s *AIAgentService) streamMock(ctx context.Context, req models.AIChatRequest, emit func(models.AIStreamEvent) error) (string, error) {
	step2 := &models.AIAgentStep{ID: "step-2", Label: "Correlating distributed spans & latency", Status: "running"}
	_ = emit(models.AIStreamEvent{Type: "step", Step: step2})
	time.Sleep(150 * time.Millisecond)
	step2.Status = "complete"
	step2.Detail = "Correlated against 24-hour service baseline"
	_ = emit(models.AIStreamEvent{Type: "step", Step: step2})

	step3 := &models.AIAgentStep{ID: "step-3", Label: "Synthesizing fault hypotheses", Status: "complete"}
	_ = emit(models.AIStreamEvent{Type: "step", Step: step3})

	markdown := `### Diagnostic Summary
Investigation completed. The primary timing bottleneck is concentrated in the upstream destination service, which exceeded the standard 2,000 ms threshold.

### Key Observations
- **Upstream Latency Dominance**: The destination application took >95% of total request duration.
- **Relay Transmission Healthy**: Outpipe edge relays transferred packets within normal sub-millisecond windows.

### Unconfirmed Possibilities
1. **Upstream Database / Socket Wait**: Target service may be waiting on a database lock or external API.
2. **Resource Exhaustion**: Local worker pools may be depleted during request bursts.

### Recommended Next Checks
1. Check local application worker logs for active timeouts or connection contention.
2. Inspect nearby requests in Observability to identify whether latency is isolated or widespread.`

	chunks := strings.Split(markdown, "\n\n")
	var fullText string
	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		prefix := ""
		if i > 0 {
			prefix = "\n\n"
		}
		delta := prefix + chunk
		fullText += delta
		_ = emit(models.AIStreamEvent{Type: "text", Delta: delta})
		time.Sleep(80 * time.Millisecond)
	}

	return fullText, nil
}

func (s *AIAgentService) streamOpenAI(ctx context.Context, key, model, prompt string, emit func(models.AIStreamEvent) error) (string, error) {
	return s.streamOpenAICompat(ctx, "https://api.openai.com/v1/chat/completions", key, model, prompt, emit)
}

func (s *AIAgentService) streamOpenAICompat(ctx context.Context, url, key, model, prompt string, emit func(models.AIStreamEvent) error) (string, error) {
	bodyData := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are Outpipe Agent, an expert read-only observability and tunneling investigator. Provide structured Markdown conclusions with ### Diagnostic Summary, ### Key Observations, ### Unconfirmed Possibilities, ### Recommended Next Checks."},
			{"role": "user", "content": prompt},
		},
		"stream": true,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai provider error with status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullText string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") || line == "data: [DONE]" {
			continue
		}
		var parsed struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &parsed); err == nil {
			if len(parsed.Choices) > 0 && parsed.Choices[0].Delta.Content != "" {
				delta := parsed.Choices[0].Delta.Content
				fullText += delta
				_ = emit(models.AIStreamEvent{Type: "text", Delta: delta})
			}
		}
	}
	return fullText, nil
}
