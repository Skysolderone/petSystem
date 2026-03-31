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

type NotificationHandler struct {
	notifications *service.NotificationService
	wsHub         *ws.Hub
	upgrader      websocket.Upgrader
}

func NewNotificationHandler(notifications *service.NotificationService, hub *ws.Hub) *NotificationHandler {
	return &NotificationHandler{
		notifications: notifications,
		wsHub:         hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

func (h *NotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, normalizedPage, normalizedPageSize, err := h.notifications.List(c.Request.Context(), middleware.MustUserID(c), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	resolved := make([]dto.NotificationResponse, 0, len(items))
	for _, item := range items {
		itemCopy := item
		resolved = append(resolved, dto.ToNotificationResponse(&itemCopy))
	}

	response.Success(c, http.StatusOK, resolved, gin.H{
		"page":      normalizedPage,
		"page_size": normalizedPageSize,
		"total":     total,
	})
}

func (h *NotificationHandler) Read(c *gin.Context) {
	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, badRequest("invalid notification id"))
		return
	}

	notification, unreadCount, err := h.notifications.MarkRead(c.Request.Context(), middleware.MustUserID(c), notificationID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.NotificationReadResponse{
		ID:          notification.ID.String(),
		Read:        true,
		UnreadCount: unreadCount,
	}, nil)
}

func (h *NotificationHandler) ReadAll(c *gin.Context) {
	updated, unreadCount, err := h.notifications.MarkAllRead(c.Request.Context(), middleware.MustUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.NotificationReadAllResponse{
		Updated:     updated,
		UnreadCount: unreadCount,
	}, nil)
}

func (h *NotificationHandler) RegisterPushToken(c *gin.Context) {
	var req dto.RegisterPushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid push token payload"))
		return
	}

	pushToken, err := h.notifications.RegisterPushToken(c.Request.Context(), middleware.MustUserID(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, dto.ToPushTokenResponse(pushToken), nil)
}

func (h *NotificationHandler) UnregisterPushToken(c *gin.Context) {
	var req dto.UnregisterPushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, badRequest("invalid push token payload"))
		return
	}

	if err := h.notifications.UnregisterPushToken(c.Request.Context(), middleware.MustUserID(c), req.Token); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true}, nil)
}

func (h *NotificationHandler) Stream(c *gin.Context) {
	userID := middleware.MustUserID(c)
	notifications, unreadCount, err := h.notifications.Snapshot(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.wsHub.SubscribeUser(userID, conn)
	defer h.wsHub.UnsubscribeUser(userID, conn)

	items := make([]dto.NotificationResponse, 0, len(notifications))
	for _, item := range notifications {
		itemCopy := item
		items = append(items, dto.ToNotificationResponse(&itemCopy))
	}
	_ = conn.WriteJSON(dto.NotificationStreamMessage{
		Type:          "notification_sync",
		Timestamp:     time.Now(),
		Notifications: items,
		UnreadCount:   unreadCount,
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
