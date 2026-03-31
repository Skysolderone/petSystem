package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/ai"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/pagination"
)

type HealthService struct {
	pets    healthPetRepository
	records healthRepository
	devices healthDeviceDataRepository
	ai      *ai.HealthAI
	llm     *ai.Assistant
}

type healthPetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

type healthRepository interface {
	CreateRecord(ctx context.Context, record *model.HealthRecord) error
	ListRecordsByPet(ctx context.Context, petID uuid.UUID, page, pageSize int, typeFilter string) ([]model.HealthRecord, int64, error)
	ListRecentRecordsByPet(ctx context.Context, petID uuid.UUID, limit int) ([]model.HealthRecord, error)
	GetRecordByID(ctx context.Context, petID, id uuid.UUID) (*model.HealthRecord, error)
	UpdateRecord(ctx context.Context, record *model.HealthRecord) error
	DeleteRecord(ctx context.Context, record *model.HealthRecord) error
	CreateAlert(ctx context.Context, alert *model.HealthAlert) error
	ListAlertsByPet(ctx context.Context, petID uuid.UUID) ([]model.HealthAlert, error)
}

type healthDeviceDataRepository interface {
	ListDataPointsByPet(ctx context.Context, petID uuid.UUID, limit int) ([]model.DeviceDataPoint, error)
}

func NewHealthService(
	pets healthPetRepository,
	records healthRepository,
	devices healthDeviceDataRepository,
	healthAI *ai.HealthAI,
) *HealthService {
	return NewHealthServiceWithOptions(pets, records, devices, healthAI)
}

type HealthServiceOption func(*HealthService)

func WithHealthAssistant(assistant *ai.Assistant) HealthServiceOption {
	return func(service *HealthService) {
		service.llm = assistant
	}
}

func NewHealthServiceWithOptions(
	pets healthPetRepository,
	records healthRepository,
	devices healthDeviceDataRepository,
	healthAI *ai.HealthAI,
	options ...HealthServiceOption,
) *HealthService {
	service := &HealthService{
		pets:    pets,
		records: records,
		devices: devices,
		ai:      healthAI,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *HealthService) ListRecords(ctx context.Context, ownerID, petID uuid.UUID, page, pageSize int, typeFilter string) ([]model.HealthRecord, int64, int, int, error) {
	if _, err := s.petForOwner(ctx, ownerID, petID); err != nil {
		return nil, 0, 0, 0, err
	}

	page, pageSize = pagination.Normalize(page, pageSize)
	records, total, err := s.records.ListRecordsByPet(ctx, petID, page, pageSize, typeFilter)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_health_records_failed", "failed to load health records", err)
	}
	return records, total, page, pageSize, nil
}

func (s *HealthService) CreateRecord(ctx context.Context, ownerID, petID uuid.UUID, req dto.CreateHealthRecordRequest) (*model.HealthRecord, error) {
	if _, err := s.petForOwner(ctx, ownerID, petID); err != nil {
		return nil, err
	}

	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	record := &model.HealthRecord{
		PetID:       petID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Data:        datatypes.JSON(dto.EncodeMap(req.Data)),
		DueDate:     req.DueDate,
		Attachments: datatypes.JSON(dto.EncodeStringArray(req.Attachments)),
		RecordedAt:  recordedAt,
	}

	if err := s.records.CreateRecord(ctx, record); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_health_record_failed", "failed to create health record", err)
	}

	if alert := buildHealthAlert(record); alert != nil {
		_ = s.records.CreateAlert(ctx, alert)
	}

	return record, nil
}

func (s *HealthService) GetRecord(ctx context.Context, ownerID, petID, recordID uuid.UUID) (*model.HealthRecord, error) {
	if _, err := s.petForOwner(ctx, ownerID, petID); err != nil {
		return nil, err
	}

	record, err := s.records.GetRecordByID(ctx, petID, recordID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_health_record_failed", "failed to load health record", err)
	}
	if record == nil {
		return nil, apperror.New(http.StatusNotFound, "health_record_not_found", "health record not found")
	}
	return record, nil
}

func (s *HealthService) UpdateRecord(ctx context.Context, ownerID, petID, recordID uuid.UUID, req dto.UpdateHealthRecordRequest) (*model.HealthRecord, error) {
	record, err := s.GetRecord(ctx, ownerID, petID, recordID)
	if err != nil {
		return nil, err
	}

	if req.Type != nil {
		record.Type = *req.Type
	}
	if req.Title != nil {
		record.Title = *req.Title
	}
	if req.Description != nil {
		record.Description = *req.Description
	}
	if req.Data != nil {
		record.Data = datatypes.JSON(dto.EncodeMap(req.Data))
	}
	if req.DueDate != nil {
		record.DueDate = req.DueDate
	}
	if req.Attachments != nil {
		record.Attachments = datatypes.JSON(dto.EncodeStringArray(req.Attachments))
	}
	if req.RecordedAt != nil {
		record.RecordedAt = *req.RecordedAt
	}

	if err := s.records.UpdateRecord(ctx, record); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_health_record_failed", "failed to update health record", err)
	}

	return record, nil
}

func (s *HealthService) DeleteRecord(ctx context.Context, ownerID, petID, recordID uuid.UUID) error {
	record, err := s.GetRecord(ctx, ownerID, petID, recordID)
	if err != nil {
		return err
	}

	if err := s.records.DeleteRecord(ctx, record); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_health_record_failed", "failed to delete health record", err)
	}
	return nil
}

func (s *HealthService) Alerts(ctx context.Context, ownerID, petID uuid.UUID) ([]model.HealthAlert, error) {
	if _, err := s.petForOwner(ctx, ownerID, petID); err != nil {
		return nil, err
	}

	alerts, err := s.records.ListAlertsByPet(ctx, petID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_health_alerts_failed", "failed to load health alerts", err)
	}
	return alerts, nil
}

func (s *HealthService) Summary(ctx context.Context, ownerID, petID uuid.UUID) (ai.HealthSummary, error) {
	pet, err := s.petForOwner(ctx, ownerID, petID)
	if err != nil {
		return ai.HealthSummary{}, err
	}

	records, err := s.records.ListRecentRecordsByPet(ctx, petID, 10)
	if err != nil {
		return ai.HealthSummary{}, apperror.Wrap(http.StatusInternalServerError, "list_health_records_failed", "failed to load health records", err)
	}
	alerts, err := s.records.ListAlertsByPet(ctx, petID)
	if err != nil {
		return ai.HealthSummary{}, apperror.Wrap(http.StatusInternalServerError, "list_health_alerts_failed", "failed to load health alerts", err)
	}
	dataPoints, err := s.devices.ListDataPointsByPet(ctx, petID, 20)
	if err != nil {
		return ai.HealthSummary{}, apperror.Wrap(http.StatusInternalServerError, "list_device_data_failed", "failed to load device data", err)
	}

	return s.ai.Summarize(*pet, records, alerts, dataPoints), nil
}

func (s *HealthService) AskAI(ctx context.Context, ownerID, petID uuid.UUID, question string) (string, error) {
	summary, err := s.Summary(ctx, ownerID, petID)
	if err != nil {
		return "", err
	}
	if s.llm != nil {
		answer, err := s.llm.AnswerHealth(ctx, question, summary)
		if err == nil && answer != "" {
			return answer, nil
		}
	}
	return s.ai.Answer(question, summary), nil
}

func (s *HealthService) petForOwner(ctx context.Context, ownerID, petID uuid.UUID) (*model.Pet, error) {
	pet, err := s.pets.GetByIDAndOwner(ctx, petID, ownerID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
	}
	if pet == nil {
		return nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
	}
	return pet, nil
}

func buildHealthAlert(record *model.HealthRecord) *model.HealthAlert {
	loweredDescription := strings.ToLower(record.Description)
	switch {
	case record.DueDate != nil && record.DueDate.Before(time.Now()):
		return &model.HealthAlert{
			PetID:     record.PetID,
			AlertType: "reminder",
			Severity:  "medium",
			Title:     record.Title + " 已到期",
			Message:   "该健康事项已超过计划日期，建议尽快处理。",
			Source:    "manual",
		}
	case record.Type == "symptom" && containsUrgentKeyword(loweredDescription):
		return &model.HealthAlert{
			PetID:     record.PetID,
			AlertType: "ai_warning",
			Severity:  "high",
			Title:     "症状记录需要重点关注",
			Message:   "检测到呕吐、腹泻、出血或抽搐等高风险关键词，建议尽快咨询兽医。",
			Source:    "ai",
		}
	default:
		return nil
	}
}

func containsUrgentKeyword(text string) bool {
	keywords := []string{"vomit", "seizure", "bleed", "diarrhea", "呕吐", "腹泻", "抽搐", "出血"}
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
