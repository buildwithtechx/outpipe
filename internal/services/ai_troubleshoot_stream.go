package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

func (s *AIAgentService) streamMock(ctx context.Context, _ models.AIChatRequest, emit func(models.AIStreamEvent) error) (string, error) {
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

func (s *AIAgentService) streamOpenAI(ctx context.Context, key, model string, messages []map[string]string, emit func(models.AIStreamEvent) error) (string, error) {
	return s.streamOpenAICompat(ctx, "https://api.openai.com/v1/chat/completions", key, model, messages, emit)
}

func (s *AIAgentService) streamOpenAICompat(ctx context.Context, url, key, model string, messages []map[string]string, emit func(models.AIStreamEvent) error) (string, error) {
	bodyData := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
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
	if err := scanner.Err(); err != nil {
		return fullText, fmt.Errorf("read ai stream: %w", err)
	}
	return fullText, nil
}

func (s *AIAgentService) streamAnthropic(ctx context.Context, key, model, systemPrompt string, messages []map[string]string, emit func(models.AIStreamEvent) error) (string, error) {
	bodyData := map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages":   messages,
		"stream":     true,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("marshal anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create anthropic request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do anthropic request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic provider error with status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullText string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var parsed struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &parsed); err == nil {
			if parsed.Type == "content_block_delta" && parsed.Delta.Text != "" {
				delta := parsed.Delta.Text
				fullText += delta
				_ = emit(models.AIStreamEvent{Type: "text", Delta: delta})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fullText, fmt.Errorf("read anthropic stream: %w", err)
	}
	return fullText, nil
}
