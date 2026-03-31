package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"petverse/server/internal/dto"
	"petverse/server/internal/middleware"
	"petverse/server/internal/pkg/response"
	"petverse/server/internal/service"
	"petverse/server/internal/ws"
)

type DeviceHandler struct {
	devices  *service.DeviceService
	wsHub    *ws.Hub
	upgrader websocket.Upgrader
}

func NewDeviceHandler(devices *service.DeviceService, hub *ws.Hub) *DeviceHandler {
	return &DeviceHandler{
		devices: devices,
		wsHub:   hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	devices, total, normalizedPage, normalizedPageSize, err := h.devices.List(c.Request.Context(), middleware.MustUserID(c), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.DeviceResponse, 0, len(devices))
	for _, device := range devices {
		deviceCopy := device
		items = append(items, dto.ToDeviceResponse(&deviceCopy))
	}

	response.Success(c, http.StatusOK, items, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var req dto.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid device payload"))
		return
	}

	device, err := h.devices.Create(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusCreated, dto.ToDeviceResponse(device), nil)
}

func (h *DeviceHandler) Get(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	device, err := h.devices.Get(c.Request.Context(), middleware.MustUserID(c), deviceID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToDeviceResponse(device), nil)
}

func (h *DeviceHandler) Update(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	var req dto.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid device payload"))
		return
	}

	device, err := h.devices.Update(c.Request.Context(), middleware.MustUserID(c), deviceID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToDeviceResponse(device), nil)
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	if err := h.devices.Delete(c.Request.Context(), middleware.MustUserID(c), deviceID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *DeviceHandler) Command(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	var req dto.DeviceCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid device command payload"))
		return
	}

	point, err := h.devices.Command(c.Request.Context(), middleware.MustUserID(c), deviceID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, dto.DeviceCommandResponse{
		Accepted:   true,
		Command:    req.Command,
		ExecutedAt: time.Now(),
		DataPoint:  dto.ToDeviceDataPointResponse(point),
	}, nil)
}

func (h *DeviceHandler) Data(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "48"))
	points, err := h.devices.Data(c.Request.Context(), middleware.MustUserID(c), deviceID, c.Query("metric"), hours, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.DeviceDataPointResponse, 0, len(points))
	for _, point := range points {
		pointCopy := point
		items = append(items, dto.ToDeviceDataPointResponse(&pointCopy))
	}
	response.Success(c, http.StatusOK, items, nil)
}

func (h *DeviceHandler) Status(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	device, points, err := h.devices.Status(c.Request.Context(), middleware.MustUserID(c), deviceID)
	if err != nil {
		response.Error(c, err)
		return
	}

	items := make([]dto.DeviceDataPointResponse, 0, len(points))
	for _, point := range points {
		pointCopy := point
		items = append(items, dto.ToDeviceDataPointResponse(&pointCopy))
	}

	response.Success(c, http.StatusOK, dto.DeviceStatusResponse{
		Device:           dto.ToDeviceResponse(device),
		LatestDataPoints: items,
	}, nil)
}

func (h *DeviceHandler) Stream(c *gin.Context) {
	deviceID, ok := parseDeviceID(c)
	if !ok {
		return
	}

	device, points, err := h.devices.Status(c.Request.Context(), middleware.MustUserID(c), deviceID)
	if err != nil {
		response.Error(c, err)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.wsHub.Subscribe(deviceID, conn)
	defer h.wsHub.Unsubscribe(deviceID, conn)

	initialPoints := make([]dto.DeviceDataPointResponse, 0, len(points))
	for _, point := range points {
		pointCopy := point
		initialPoints = append(initialPoints, dto.ToDeviceDataPointResponse(&pointCopy))
	}
	_ = conn.WriteJSON(dto.DeviceStreamMessage{
		Type:      "device_status",
		Timestamp: time.Now(),
		Status: &dto.DeviceStatusResponse{
			Device:           dto.ToDeviceResponse(device),
			LatestDataPoints: initialPoints,
		},
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func parseDeviceID(c *gin.Context) (uuid.UUID, bool) {
	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid device id"))
		return uuid.Nil, false
	}
	return deviceID, true
}
