package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"petverse/server/internal/model"
)

type Assistant struct {
	generator TextGenerator
}

type TrainingPlanDraft struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Steps       []map[string]any `json:"steps"`
}

func NewAssistant(generator TextGenerator) *Assistant {
	if generator == nil {
		return nil
	}
	return &Assistant{generator: generator}
}

func (a *Assistant) AnswerHealth(ctx context.Context, question string, summary HealthSummary) (string, error) {
	if a == nil || a.generator == nil {
		return "", errors.New("assistant not configured")
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}

	return a.generator.Generate(
		ctx,
		"You are a pet health assistant. Respond clearly, cautiously, and never claim to replace a veterinarian. Keep the answer under 180 words.",
		fmt.Sprintf("Question:\n%s\n\nHealth summary JSON:\n%s", question, string(summaryJSON)),
	)
}

func (a *Assistant) AnswerCommunity(ctx context.Context, question string) (string, error) {
	if a == nil || a.generator == nil {
		return "", errors.New("assistant not configured")
	}

	return a.generator.Generate(
		ctx,
		"You help pet owners choose services and evaluate community advice. Be practical, concise, and safety-aware.",
		fmt.Sprintf("Answer this pet community question in a concise way:\n%s", question),
	)
}

func (a *Assistant) GenerateTrainingPlan(ctx context.Context, pet model.Pet, goal, category, difficulty string) (TrainingPlanDraft, error) {
	if a == nil || a.generator == nil {
		return TrainingPlanDraft{}, errors.New("assistant not configured")
	}

	prompt := fmt.Sprintf(
		"Create a pet training plan as JSON only. Pet name: %s. Species: %s. Breed: %s. Goal: %s. Category: %s. Difficulty: %s.\n"+
			"Return a JSON object with fields title, description, steps. steps must be an array of 4 to 6 objects, each containing day, title, instruction, duration_minutes, category, difficulty.",
		pet.Name,
		pet.Species,
		pet.Breed,
		goal,
		category,
		difficulty,
	)

	response, err := a.generator.Generate(
		ctx,
		"You are a pet training planner. Output valid JSON only with no markdown fences or commentary.",
		prompt,
	)
	if err != nil {
		return TrainingPlanDraft{}, err
	}

	parsed, err := parseTrainingPlanDraft(response)
	if err != nil {
		return TrainingPlanDraft{}, err
	}
	return parsed, nil
}

func parseTrainingPlanDraft(raw string) (TrainingPlanDraft, error) {
	var draft TrainingPlanDraft
	trimmed := strings.TrimSpace(raw)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		trimmed = trimmed[start : end+1]
	}
	if err := json.Unmarshal([]byte(trimmed), &draft); err != nil {
		return TrainingPlanDraft{}, err
	}
	if draft.Title == "" || len(draft.Steps) == 0 {
		return TrainingPlanDraft{}, errors.New("training plan draft missing required fields")
	}
	return draft, nil
}
