package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ExpoDefaultURL = "https://exp.host/--/api/v2/push/send"

type Dispatcher interface {
	Send(ctx context.Context, tokens []string, payload Payload) ([]string, error)
}

type Config struct {
	Provider    string
	ExpoURL     string
	AccessToken string
	Timeout     time.Duration
}

type Payload struct {
	Title string
	Body  string
	Data  map[string]any
}

func NewDispatcher(cfg Config) Dispatcher {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "none" {
		return nil
	}

	switch provider {
	case "expo":
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		endpoint := strings.TrimSpace(cfg.ExpoURL)
		if endpoint == "" {
			endpoint = ExpoDefaultURL
		}
		return &ExpoDispatcher{
			endpoint:    endpoint,
			accessToken: strings.TrimSpace(cfg.AccessToken),
			httpClient:  &http.Client{Timeout: timeout},
		}
	default:
		return nil
	}
}

type ExpoDispatcher struct {
	endpoint    string
	accessToken string
	httpClient  *http.Client
}

func (d *ExpoDispatcher) Send(ctx context.Context, tokens []string, payload Payload) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	messages := make([]map[string]any, 0, len(tokens))
	for _, token := range tokens {
		if token == "" {
			continue
		}
		messages = append(messages, map[string]any{
			"to":    token,
			"title": payload.Title,
			"body":  payload.Body,
			"data":  payload.Data,
		})
	}
	if len(messages) == 0 {
		return nil, nil
	}

	raw, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	if d.accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+d.accessToken)
	}

	response, err := d.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("expo push send failed: %s", strings.TrimSpace(string(body)))
	}

	var envelope struct {
		Data []struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Details struct {
				Error string `json:"error"`
			} `json:"details"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	invalidTokens := make([]string, 0)
	for index, item := range envelope.Data {
		if item.Status == "error" && item.Details.Error == "DeviceNotRegistered" && index < len(messages) {
			token, _ := messages[index]["to"].(string)
			if token != "" {
				invalidTokens = append(invalidTokens, token)
			}
		}
	}

	return invalidTokens, nil
}
