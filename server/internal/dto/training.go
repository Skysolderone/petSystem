package dto

import (
	"time"

	"petverse/server/internal/model"
)

type CreateTrainingPlanRequest struct {
	Title       string           `json:"title" binding:"required,min=1,max=100"`
	Description string           `json:"description"`
	Difficulty  string           `json:"difficulty"`
	Category    string           `json:"category"`
	Steps       []map[string]any `json:"steps"`
}

type GenerateTrainingPlanRequest struct {
	Goal       string `json:"goal" binding:"required,min=2,max=100"`
	Difficulty string `json:"difficulty"`
	Category   string `json:"category"`
}

type UpdateTrainingPlanRequest struct {
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	Difficulty  *string           `json:"difficulty"`
	Category    *string           `json:"category"`
	Steps       *[]map[string]any `json:"steps"`
	Progress    *int              `json:"progress"`
}

type TrainingPlanResponse struct {
	ID          string           `json:"id"`
	PetID       string           `json:"pet_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Difficulty  string           `json:"difficulty"`
	Category    string           `json:"category"`
	Steps       []map[string]any `json:"steps"`
	AIGenerated bool             `json:"ai_generated"`
	Progress    int              `json:"progress"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func ToTrainingPlanResponse(plan *model.TrainingPlan) TrainingPlanResponse {
	return TrainingPlanResponse{
		ID:          plan.ID.String(),
		PetID:       plan.PetID.String(),
		Title:       plan.Title,
		Description: plan.Description,
		Difficulty:  plan.Difficulty,
		Category:    plan.Category,
		Steps:       decodeMapSlice(plan.Steps),
		AIGenerated: plan.AIGenerated,
		Progress:    plan.Progress,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
	}
}
