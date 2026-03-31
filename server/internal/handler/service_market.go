package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type ServiceMarketHandler struct {
	services *service.ServiceMarketService
}

func NewServiceMarketHandler(services *service.ServiceMarketService) *ServiceMarketHandler {
	return &ServiceMarketHandler{services: services}
}

func (h *ServiceMarketHandler) List(c *gin.Context) {
	var lat, lng *float64
	if value := c.Query("lat"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			lat = &parsed
		}
	}
	if value := c.Query("lng"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			lng = &parsed
		}
	}

	providers, distances, err := h.services.List(c.Request.Context(), lat, lng, c.Query("type"))
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.ServiceProviderResponse, 0, len(providers))
	for _, provider := range providers {
		providerCopy := provider
		item := dto.ToServiceProviderResponse(&providerCopy)
		if distance, ok := distances[provider.ID]; ok {
			item.DistanceKm = distance
		}
		items = append(items, item)
	}

	response.Success(c, http.StatusOK, items, nil)
}

func (h *ServiceMarketHandler) Get(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid provider id"))
		return
	}

	provider, err := h.services.Get(c.Request.Context(), providerID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToServiceProviderResponse(provider), nil)
}

func (h *ServiceMarketHandler) Reviews(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid provider id"))
		return
	}

	reviews, err := h.services.Reviews(c.Request.Context(), providerID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.BookingResponse, 0, len(reviews))
	for _, review := range reviews {
		reviewCopy := review
		items = append(items, dto.ToBookingResponse(&reviewCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *ServiceMarketHandler) Availability(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid provider id"))
		return
	}

	slots, err := h.services.Availability(c.Request.Context(), providerID)
	if err != nil {
		response.Error(c, err)
		return
	}

	formattedSlots := make([]string, 0, len(slots))
	for _, slot := range slots {
		formattedSlots = append(formattedSlots, slot.Format(time.RFC3339))
	}

	response.Success(c, http.StatusOK, dto.ServiceAvailabilityResponse{
		ProviderID: providerID.String(),
		Slots:      formattedSlots,
	}, nil)
}
