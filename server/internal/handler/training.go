package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type TrainingHandler struct {
	training *service.TrainingService
}

func NewTrainingHandler(training *service.TrainingService) *TrainingHandler {
	return &TrainingHandler{training: training}
}

func (h *TrainingHandler) List(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}
	plans, err := h.training.List(c.Request.Context(), middleware.MustUserID(c), petID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.TrainingPlanResponse, 0, len(plans))
	for _, plan := range plans {
		planCopy := plan
		items = append(items, dto.ToTrainingPlanResponse(&planCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *TrainingHandler) Create(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	var req dto.CreateTrainingPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid training payload"))
		return
	}

	plan, err := h.training.Create(c.Request.Context(), middleware.MustUserID(c), petID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.ToTrainingPlanResponse(plan), nil)
}

func (h *TrainingHandler) Generate(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	var req dto.GenerateTrainingPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid training generation payload"))
		return
	}

	plan, err := h.training.Generate(c.Request.Context(), middleware.MustUserID(c), petID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.ToTrainingPlanResponse(plan), nil)
}

func (h *TrainingHandler) Get(c *gin.Context) {
	planID, ok := parseTrainingID(c)
	if !ok {
		return
	}
	plan, err := h.training.Get(c.Request.Context(), middleware.MustUserID(c), planID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToTrainingPlanResponse(plan), nil)
}

func (h *TrainingHandler) Update(c *gin.Context) {
	planID, ok := parseTrainingID(c)
	if !ok {
		return
	}

	var req dto.UpdateTrainingPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid training payload"))
		return
	}

	plan, err := h.training.Update(c.Request.Context(), middleware.MustUserID(c), planID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToTrainingPlanResponse(plan), nil)
}

func (h *TrainingHandler) Delete(c *gin.Context) {
	planID, ok := parseTrainingID(c)
	if !ok {
		return
	}

	if err := h.training.Delete(c.Request.Context(), middleware.MustUserID(c), planID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func parseTrainingID(c *gin.Context) (uuid.UUID, bool) {
	planID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid training plan id"))
		return uuid.Nil, false
	}
	return planID, true
}
