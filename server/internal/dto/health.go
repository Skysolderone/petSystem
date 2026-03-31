package dto

import (
	"encoding/json"
	"time"

	"petverse/server/internal/model"
	"petverse/server/internal/pkg/ai"
)

type CreateHealthRecordRequest struct {
	Type        string         `json:"type" binding:"required,min=1,max=30"`
	Title       string         `json:"title" binding:"required,min=1,max=200"`
	Description string         `json:"description"`
	Data        map[string]any `json:"data"`
	DueDate     *time.Time     `json:"due_date"`
	Attachments []string       `json:"attachments"`
	RecordedAt  *time.Time     `json:"recorded_at"`
}

type UpdateHealthRecordRequest struct {
	Type        *string        `json:"type"`
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Data        map[string]any `json:"data"`
	DueDate     *time.Time     `json:"due_date"`
	Attachments []string       `json:"attachments"`
	RecordedAt  *time.Time     `json:"recorded_at"`
}

type AskAIRequest struct {
	Question string `json:"question" binding:"required,min=3,max=500"`
}

type HealthRecordResponse struct {
	ID          string         `json:"id"`
	PetID       string         `json:"pet_id"`
	Type        string         `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Data        map[string]any `json:"data"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	Attachments []string       `json:"attachments"`
	RecordedAt  time.Time      `json:"recorded_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type HealthAlertResponse struct {
	ID          string    `json:"id"`
	PetID       string    `json:"pet_id"`
	AlertType   string    `json:"alert_type"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Source      string    `json:"source"`
	IsRead      bool      `json:"is_read"`
	IsDismissed bool      `json:"is_dismissed"`
	CreatedAt   time.Time `json:"created_at"`
}

type HealthSummaryResponse struct {
	Score              int       `json:"score"`
	Status             string    `json:"status"`
	Insights           []string  `json:"insights"`
	RecommendedActions []string  `json:"recommended_actions"`
	DataPointsAnalyzed int       `json:"data_points_analyzed"`
	GeneratedAt        time.Time `json:"generated_at"`
}

type HealthAIAnswerResponse struct {
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

func ToHealthRecordResponse(record *model.HealthRecord) HealthRecordResponse {
	return HealthRecordResponse{
		ID:          record.ID.String(),
		PetID:       record.PetID.String(),
		Type:        record.Type,
		Title:       record.Title,
		Description: record.Description,
		Data:        decodeMap(record.Data),
		DueDate:     record.DueDate,
		Attachments: decodeStringArray(record.Attachments),
		RecordedAt:  record.RecordedAt,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func ToHealthAlertResponse(alert *model.HealthAlert) HealthAlertResponse {
	return HealthAlertResponse{
		ID:          alert.ID.String(),
		PetID:       alert.PetID.String(),
		AlertType:   alert.AlertType,
		Severity:    alert.Severity,
		Title:       alert.Title,
		Message:     alert.Message,
		Source:      alert.Source,
		IsRead:      alert.IsRead,
		IsDismissed: alert.IsDismissed,
		CreatedAt:   alert.CreatedAt,
	}
}

func ToHealthSummaryResponse(summary ai.HealthSummary) HealthSummaryResponse {
	return HealthSummaryResponse{
		Score:              summary.Score,
		Status:             summary.Status,
		Insights:           summary.Insights,
		RecommendedActions: summary.RecommendedActions,
		DataPointsAnalyzed: summary.DataPointsAnalyzed,
		GeneratedAt:        summary.GeneratedAt,
	}
}

func EncodeMap(payload map[string]any) []byte {
	if len(payload) == 0 {
		return []byte("{}")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

func DecodeJSONMap(raw []byte) map[string]any {
	return decodeMap(raw)
}

func decodeMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return map[string]any{}
	}
	return result
}
