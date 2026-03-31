package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
)

type BookingHandler struct {
	bookings *service.BookingService
}

func NewBookingHandler(bookings *service.BookingService) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

func (h *BookingHandler) Create(c *gin.Context) {
	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid booking payload"))
		return
	}

	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		response.Error(c, badRequest("invalid provider id"))
		return
	}

	booking, err := h.bookings.Create(c.Request.Context(), middleware.MustUserID(c), providerID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToBookingResponse(booking), nil)
}

func (h *BookingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	bookings, total, normalizedPage, normalizedPageSize, err := h.bookings.List(c.Request.Context(), middleware.MustUserID(c), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.BookingResponse, 0, len(bookings))
	for _, booking := range bookings {
		bookingCopy := booking
		items = append(items, dto.ToBookingResponse(&bookingCopy))
	}
	response.Success(c, http.StatusOK, items, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *BookingHandler) Get(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid booking id"))
		return
	}

	booking, err := h.bookings.Get(c.Request.Context(), middleware.MustUserID(c), bookingID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToBookingResponse(booking), nil)
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid booking id"))
		return
	}

	var req dto.CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid cancel payload"))
		return
	}

	booking, err := h.bookings.Cancel(c.Request.Context(), middleware.MustUserID(c), bookingID, req.Reason)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToBookingResponse(booking), nil)
}

func (h *BookingHandler) Review(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid booking id"))
		return
	}

	var req dto.ReviewBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid review payload"))
		return
	}

	booking, err := h.bookings.Review(c.Request.Context(), middleware.MustUserID(c), bookingID, req.Rating, req.Review)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToBookingResponse(booking), nil)
}
