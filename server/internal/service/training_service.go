package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/ai"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/events"
)

type TrainingService struct {
	plans    trainingRepository
	pets     trainingPetRepository
	notifier trainingNotifier
	events   events.Publisher
	llm      *ai.Assistant
}

type trainingRepository interface {
	Create(ctx context.Context, plan *model.TrainingPlan) error
	ListByPet(ctx context.Context, petID uuid.UUID) ([]model.TrainingPlan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.TrainingPlan, error)
	Update(ctx context.Context, plan *model.TrainingPlan) error
	Delete(ctx context.Context, plan *model.TrainingPlan) error
}

type trainingPetRepository interface {
	GetByIDAndOwner(ctx context.Context, id, ownerID uuid.UUID) (*model.Pet, error)
}

type trainingNotifier interface {
	Create(ctx context.Context, userID uuid.UUID, input NotificationCreateInput) (*model.Notification, error)
}

type TrainingServiceOption func(*TrainingService)

func WithTrainingAssistant(assistant *ai.Assistant) TrainingServiceOption {
	return func(service *TrainingService) {
		service.llm = assistant
	}
}

func WithTrainingEvents(publisher events.Publisher) TrainingServiceOption {
	return func(service *TrainingService) {
		service.events = publisher
	}
}

func NewTrainingService(plans trainingRepository, pets trainingPetRepository, notifier trainingNotifier, options ...TrainingServiceOption) *TrainingService {
	service := &TrainingService{
		plans:    plans,
		pets:     pets,
		notifier: notifier,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *TrainingService) List(ctx context.Context, userID, petID uuid.UUID) ([]model.TrainingPlan, error) {
	if _, err := s.getOwnedPet(ctx, userID, petID); err != nil {
		return nil, err
	}
	plans, err := s.plans.ListByPet(ctx, petID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "list_training_plans_failed", "failed to load training plans", err)
	}
	return plans, nil
}

func (s *TrainingService) Create(ctx context.Context, userID, petID uuid.UUID, req dto.CreateTrainingPlanRequest) (*model.TrainingPlan, error) {
	pet, err := s.getOwnedPet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}

	plan := &model.TrainingPlan{
		PetID:       pet.ID,
		Title:       req.Title,
		Description: req.Description,
		Difficulty:  defaultTrainingDifficulty(req.Difficulty),
		Category:    defaultTrainingCategory(req.Category),
		Steps:       datatypes.JSON(dto.EncodeMapSlice(req.Steps)),
		AIGenerated: false,
		Progress:    0,
	}
	if err := s.plans.Create(ctx, plan); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_training_plan_failed", "failed to create training plan", err)
	}

	s.publish(ctx, "training.plan.created", map[string]any{
		"plan_id":      plan.ID.String(),
		"pet_id":       plan.PetID.String(),
		"user_id":      userID.String(),
		"title":        plan.Title,
		"difficulty":   plan.Difficulty,
		"category":     plan.Category,
		"ai_generated": plan.AIGenerated,
	})
	s.notify(ctx, userID, "training", "已创建训练计划", fmt.Sprintf("%s 的训练计划《%s》已保存。", pet.Name, plan.Title), map[string]any{"pet_id": pet.ID.String(), "plan_id": plan.ID.String()})
	return plan, nil
}

func (s *TrainingService) Generate(ctx context.Context, userID, petID uuid.UUID, req dto.GenerateTrainingPlanRequest) (*model.TrainingPlan, error) {
	pet, err := s.getOwnedPet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}

	category := defaultTrainingCategory(req.Category)
	difficulty := defaultTrainingDifficulty(req.Difficulty)
	title := fmt.Sprintf("%s 的 %s 训练计划", pet.Name, req.Goal)
	description := fmt.Sprintf("根据 %s 的品种、阶段和目标生成的 %s 训练计划。", pet.Name, category)
	steps := generateTrainingSteps(pet, req.Goal, category, difficulty)
	if s.llm != nil {
		draft, err := s.llm.GenerateTrainingPlan(ctx, *pet, req.Goal, category, difficulty)
		if err == nil {
			if draft.Title != "" {
				title = draft.Title
			}
			if draft.Description != "" {
				description = draft.Description
			}
			if len(draft.Steps) > 0 {
				steps = draft.Steps
			}
		}
	}

	plan := &model.TrainingPlan{
		PetID:       pet.ID,
		Title:       title,
		Description: description,
		Difficulty:  difficulty,
		Category:    category,
		Steps:       datatypes.JSON(dto.EncodeMapSlice(steps)),
		AIGenerated: true,
		Progress:    0,
	}
	if err := s.plans.Create(ctx, plan); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "generate_training_plan_failed", "failed to generate training plan", err)
	}

	s.publish(ctx, "training.plan.generated", map[string]any{
		"plan_id":      plan.ID.String(),
		"pet_id":       plan.PetID.String(),
		"user_id":      userID.String(),
		"title":        plan.Title,
		"goal":         req.Goal,
		"difficulty":   plan.Difficulty,
		"category":     plan.Category,
		"ai_generated": plan.AIGenerated,
	})
	s.notify(ctx, userID, "system", "AI 训练计划已生成", fmt.Sprintf("%s 的目标“%s”已经生成可执行训练步骤。", pet.Name, req.Goal), map[string]any{"pet_id": pet.ID.String(), "plan_id": plan.ID.String()})
	return plan, nil
}

func (s *TrainingService) Get(ctx context.Context, userID, planID uuid.UUID) (*model.TrainingPlan, error) {
	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "get_training_plan_failed", "failed to load training plan", err)
	}
	if plan == nil {
		return nil, apperror.New(http.StatusNotFound, "training_plan_not_found", "training plan not found")
	}
	if _, err := s.getOwnedPet(ctx, userID, plan.PetID); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *TrainingService) Update(ctx context.Context, userID, planID uuid.UUID, req dto.UpdateTrainingPlanRequest) (*model.TrainingPlan, error) {
	plan, err := s.Get(ctx, userID, planID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		plan.Title = *req.Title
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Difficulty != nil {
		plan.Difficulty = defaultTrainingDifficulty(*req.Difficulty)
	}
	if req.Category != nil {
		plan.Category = defaultTrainingCategory(*req.Category)
	}
	if req.Steps != nil {
		plan.Steps = datatypes.JSON(dto.EncodeMapSlice(*req.Steps))
	}
	if req.Progress != nil {
		progress := *req.Progress
		if progress < 0 || progress > 100 {
			return nil, apperror.New(http.StatusBadRequest, "invalid_progress", "progress must be between 0 and 100")
		}
		plan.Progress = progress
	}

	if err := s.plans.Update(ctx, plan); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "update_training_plan_failed", "failed to update training plan", err)
	}
	return plan, nil
}

func (s *TrainingService) Delete(ctx context.Context, userID, planID uuid.UUID) error {
	plan, err := s.Get(ctx, userID, planID)
	if err != nil {
		return err
	}
	if err := s.plans.Delete(ctx, plan); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "delete_training_plan_failed", "failed to delete training plan", err)
	}
	return nil
}

func (s *TrainingService) getOwnedPet(ctx context.Context, userID, petID uuid.UUID) (*model.Pet, error) {
	pet, err := s.pets.GetByIDAndOwner(ctx, petID, userID)
	if err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "pet_lookup_failed", "failed to load pet", err)
	}
	if pet == nil {
		return nil, apperror.New(http.StatusNotFound, "pet_not_found", "pet not found")
	}
	return pet, nil
}

func (s *TrainingService) notify(ctx context.Context, userID uuid.UUID, notificationType, title, body string, data map[string]any) {
	if s.notifier == nil {
		return
	}
	_, _ = s.notifier.Create(ctx, userID, NotificationCreateInput{
		Type:  notificationType,
		Title: title,
		Body:  body,
		Data:  data,
	})
}

func defaultTrainingDifficulty(value string) string {
	if value == "" {
		return "beginner"
	}
	return value
}

func defaultTrainingCategory(value string) string {
	if value == "" {
		return "obedience"
	}
	return value
}

func generateTrainingSteps(pet *model.Pet, goal, category, difficulty string) []map[string]any {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		goal = "基础服从"
	}

	baseSteps := []string{
		"用固定口令建立注意力，确认奖励物有效。",
		"把目标动作拆分成 3 个小步骤，每次训练不超过 10 分钟。",
		"动作完成后立即奖励，并记录成功率。",
		"在不同环境中重复练习，逐步降低零食依赖。",
	}

	steps := make([]map[string]any, 0, len(baseSteps))
	for index, instruction := range baseSteps {
		steps = append(steps, map[string]any{
			"day":              index + 1,
			"title":            fmt.Sprintf("第 %d 天 · %s", index+1, goal),
			"duration_minutes": 8 + index*2,
			"category":         category,
			"difficulty":       difficulty,
			"instruction":      fmt.Sprintf("%s %s 当前品种/体型建议保持节奏稳定，结束前用口令复盘。", instruction, pet.Name),
		})
	}

	steps = append(steps, map[string]any{
		"day":              len(baseSteps) + 1,
		"title":            "巩固与泛化",
		"duration_minutes": 12,
		"category":         category,
		"difficulty":       difficulty,
		"instruction":      fmt.Sprintf("在散步、客厅和进食前各完成一次 %s，确认 %s 能在分心环境里响应。", goal, pet.Name),
	})
	return steps
}

func (s *TrainingService) publish(ctx context.Context, subject string, payload any) {
	if s.events == nil {
		return
	}
	_ = s.events.PublishJSON(ctx, subject, payload)
}
