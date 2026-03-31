package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TextGenerator interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type TextGeneratorConfig struct {
	Provider    string
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
	Timeout     time.Duration
}

func NewTextGenerator(cfg TextGeneratorConfig) TextGenerator {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "local" || cfg.APIKey == "" || cfg.Model == "" || cfg.BaseURL == "" {
		return nil
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	switch provider {
	case "openai", "openai_compatible":
		return &openAICompatibleClient{
			endpoint:    appendEndpoint(cfg.BaseURL, "/chat/completions"),
			apiKey:      cfg.APIKey,
			model:       cfg.Model,
			temperature: cfg.Temperature,
			httpClient:  &http.Client{Timeout: timeout},
		}
	case "anthropic", "claude":
		return &anthropicClient{
			endpoint:    appendEndpoint(cfg.BaseURL, "/messages"),
			apiKey:      cfg.APIKey,
			model:       cfg.Model,
			temperature: cfg.Temperature,
			httpClient:  &http.Client{Timeout: timeout},
		}
	default:
		return nil
	}
}

type openAICompatibleClient struct {
	endpoint    string
	apiKey      string
	model       string
	temperature float64
	httpClient  *http.Client
}

func (c *openAICompatibleClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model":       c.model,
		"temperature": c.temperature,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("openai compatible request failed: %s", strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("openai compatible response contained no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

type anthropicClient struct {
	endpoint    string
	apiKey      string
	model       string
	temperature float64
	httpClient  *http.Client
}

func (c *anthropicClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model":       c.model,
		"max_tokens":  900,
		"temperature": c.temperature,
		"system":      systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", "2023-06-01")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("anthropic request failed: %s", strings.TrimSpace(string(body)))
	}

	var parsed struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Content) == 0 {
		return "", errors.New("anthropic response contained no content")
	}
	return strings.TrimSpace(parsed.Content[0].Text), nil
}

func appendEndpoint(baseURL, suffix string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, suffix) {
		return trimmed
	}
	return trimmed + suffix
}
