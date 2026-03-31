package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type HealthHandler struct {
	health *service.HealthService
}

func NewHealthHandler(health *service.HealthService) *HealthHandler {
	return &HealthHandler{health: health}
}

func (h *HealthHandler) List(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	typeFilter := c.Query("type")

	records, total, normalizedPage, normalizedPageSize, err := h.health.ListRecords(c.Request.Context(), middleware.MustUserID(c), petID, page, pageSize, typeFilter)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.HealthRecordResponse, 0, len(records))
	for _, record := range records {
		recordCopy := record
		items = append(items, dto.ToHealthRecordResponse(&recordCopy))
	}

	response.Success(c, http.StatusOK, items, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *HealthHandler) Create(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	var req dto.CreateHealthRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid health record payload"))
		return
	}

	record, err := h.health.CreateRecord(c.Request.Context(), middleware.MustUserID(c), petID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToHealthRecordResponse(record), nil)
}

func (h *HealthHandler) Get(c *gin.Context) {
	petID, recordID, ok := parsePetAndRecordID(c)
	if !ok {
		return
	}

	record, err := h.health.GetRecord(c.Request.Context(), middleware.MustUserID(c), petID, recordID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToHealthRecordResponse(record), nil)
}

func (h *HealthHandler) Update(c *gin.Context) {
	petID, recordID, ok := parsePetAndRecordID(c)
	if !ok {
		return
	}

	var req dto.UpdateHealthRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid health record payload"))
		return
	}

	record, err := h.health.UpdateRecord(c.Request.Context(), middleware.MustUserID(c), petID, recordID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToHealthRecordResponse(record), nil)
}

func (h *HealthHandler) Delete(c *gin.Context) {
	petID, recordID, ok := parsePetAndRecordID(c)
	if !ok {
		return
	}

	if err := h.health.DeleteRecord(c.Request.Context(), middleware.MustUserID(c), petID, recordID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *HealthHandler) Summary(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	summary, err := h.health.Summary(c.Request.Context(), middleware.MustUserID(c), petID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToHealthSummaryResponse(summary), nil)
}

func (h *HealthHandler) Alerts(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	alerts, err := h.health.Alerts(c.Request.Context(), middleware.MustUserID(c), petID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.HealthAlertResponse, 0, len(alerts))
	for _, alert := range alerts {
		alertCopy := alert
		items = append(items, dto.ToHealthAlertResponse(&alertCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *HealthHandler) AskAI(c *gin.Context) {
	petID, ok := parsePetID(c)
	if !ok {
		return
	}

	var req dto.AskAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid ai payload"))
		return
	}

	answer, err := h.health.AskAI(c.Request.Context(), middleware.MustUserID(c), petID, req.Question)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.HealthAIAnswerResponse{
		Question:  req.Question,
		Answer:    answer,
		CreatedAt: time.Now(),
	}, nil)
}

func parsePetID(c *gin.Context) (uuid.UUID, bool) {
	petParam := c.Param("petId")
	if petParam == "" {
		petParam = c.Param("id")
	}

	petID, err := uuid.Parse(petParam)
	if err != nil {
		response.Error(c, badRequest("invalid pet id"))
		return uuid.Nil, false
	}
	return petID, true
}

func parsePetAndRecordID(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	petID, ok := parsePetID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	recordID, err := uuid.Parse(c.Param("recordId"))
	if err != nil {
		response.Error(c, badRequest("invalid record id"))
		return uuid.Nil, uuid.Nil, false
	}
	return petID, recordID, true
}
