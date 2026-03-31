package push

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExpoDispatcherSend(t *testing.T) {
	t.Parallel()

	dispatcher := &ExpoDispatcher{
		endpoint:    ExpoDefaultURL,
		accessToken: "push-secret",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.URL.String() != ExpoDefaultURL {
					t.Fatalf("unexpected endpoint: %s", request.URL.String())
				}
				if got := request.Header.Get("Authorization"); got != "Bearer push-secret" {
					t.Fatalf("unexpected auth header: %s", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"data":[{"status":"ok"}]}`)),
				}, nil
			}),
		},
	}

	invalidTokens, err := dispatcher.Send(context.Background(), []string{"ExponentPushToken[abc]"}, Payload{
		Title: "PetVerse",
		Body:  "hello",
		Data:  map[string]any{"route": "/notifications"},
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(invalidTokens) != 0 {
		t.Fatalf("expected no invalid tokens, got %v", invalidTokens)
	}
}

func TestExpoDispatcherDetectsInvalidToken(t *testing.T) {
	t.Parallel()

	dispatcher := &ExpoDispatcher{
		endpoint: ExpoDefaultURL,
		httpClient: &http.Client{
			Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body: io.NopCloser(strings.NewReader(
						`{"data":[{"status":"error","details":{"error":"DeviceNotRegistered"}}]}`,
					)),
				}, nil
			}),
		},
	}

	invalidTokens, err := dispatcher.Send(context.Background(), []string{"ExponentPushToken[invalid]"}, Payload{
		Title: "PetVerse",
		Body:  "hello",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(invalidTokens) != 1 || invalidTokens[0] != "ExponentPushToken[invalid]" {
		t.Fatalf("unexpected invalid tokens: %v", invalidTokens)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
