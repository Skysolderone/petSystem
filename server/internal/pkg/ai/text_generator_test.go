package ai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"petverse/server/internal/model"
)

func TestOpenAICompatibleGenerator(t *testing.T) {
	t.Parallel()

	generator := &openAICompatibleClient{
		endpoint:    "https://example.com/chat/completions",
		apiKey:      "secret",
		model:       "demo-model",
		temperature: 0.2,
		httpClient: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.URL.Path != "/chat/completions" {
					t.Fatalf("unexpected path: %s", request.URL.Path)
				}
				if got := request.Header.Get("Authorization"); got != "Bearer secret" {
					t.Fatalf("unexpected auth header: %s", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"generated answer"}}]}`)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	answer, err := generator.Generate(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if answer != "generated answer" {
		t.Fatalf("Generate() = %q want %q", answer, "generated answer")
	}
}

func TestAssistantParsesTrainingDraft(t *testing.T) {
	t.Parallel()

	assistant := NewAssistant(fakeGenerator(`{
		"title":"Dexter Recall Plan",
		"description":"Short recall practice for the week.",
		"steps":[
			{"day":1,"title":"Name Response","instruction":"Say Dexter's name and reward eye contact.","duration_minutes":8}
		]
	}`))

	draft, err := assistant.GenerateTrainingPlan(context.Background(), modelPetForTest(), "recall", "obedience", "beginner")
	if err != nil {
		t.Fatalf("GenerateTrainingPlan() error = %v", err)
	}
	if draft.Title != "Dexter Recall Plan" {
		t.Fatalf("unexpected title: %q", draft.Title)
	}
	if len(draft.Steps) != 1 {
		t.Fatalf("unexpected steps length: %d", len(draft.Steps))
	}
}

func TestAssistantRejectsInvalidTrainingDraft(t *testing.T) {
	t.Parallel()

	assistant := NewAssistant(fakeGenerator("not-json"))
	_, err := assistant.GenerateTrainingPlan(context.Background(), modelPetForTest(), "recall", "obedience", "beginner")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

type fakeGenerator string

func (f fakeGenerator) Generate(_ context.Context, _, _ string) (string, error) {
	return string(f), nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func modelPetForTest() model.Pet {
	return model.Pet{
		Name:    "Dexter",
		Species: "dog",
		Breed:   "corgi",
	}
}
