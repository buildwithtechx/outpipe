package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"outpipe.dev/outpipe/internal/infra/httpclient"
)

type Config struct {
	URL         string
	APIKey      string
	FromAddress string
}

type Message struct {
	To      string
	Subject string
	HTML    string
}

type ZeptoClient struct {
	config Config
	client *http.Client
}

func NewZeptoClient(config Config, client *http.Client) (*ZeptoClient, error) {

	if strings.TrimSpace(config.URL) == "" || strings.TrimSpace(config.APIKey) == "" || strings.TrimSpace(config.FromAddress) == "" {
		return nil, fmt.Errorf("zepto url, api key, and from address are required")
	}

	if client == nil {
		client = httpclient.New(0)
	}

	return &ZeptoClient{config: config, client: client}, nil
}

func (c *ZeptoClient) Send(ctx context.Context, message Message) error {

	if strings.TrimSpace(message.To) == "" || strings.TrimSpace(message.Subject) == "" {
		return fmt.Errorf("mail recipient and subject are required")
	}

	body := map[string]any{
		"from":     map[string]string{"address": c.config.FromAddress},
		"to":       []map[string]any{{"email_address": map[string]string{"address": message.To}}},
		"subject":  message.Subject,
		"htmlbody": message.HTML,
	}
	encoded, err := json.Marshal(body)

	if err != nil {
		return fmt.Errorf("encode mail message: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL, bytes.NewReader(encoded))

	if err != nil {
		return fmt.Errorf("create mail request: %w", err)
	}

	request.Header.Set("Authorization", "Zoho-enczapikey "+c.config.APIKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)

	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("zepto mail returned status %d", response.StatusCode)
	}

	return nil
}
