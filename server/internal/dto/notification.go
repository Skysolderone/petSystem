package dto

import (
	"time"

	"petverse/server/internal/model"
)

type NotificationResponse struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data"`
	IsRead    bool           `json:"is_read"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type NotificationReadResponse struct {
	ID          string `json:"id"`
	Read        bool   `json:"read"`
	UnreadCount int64  `json:"unread_count"`
}

type NotificationReadAllResponse struct {
	Updated     int64 `json:"updated"`
	UnreadCount int64 `json:"unread_count"`
}

type NotificationStreamMessage struct {
	Type          string                 `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	Notification  *NotificationResponse  `json:"notification,omitempty"`
	Notifications []NotificationResponse `json:"notifications,omitempty"`
	UnreadCount   int64                  `json:"unread_count,omitempty"`
}

type RegisterPushTokenRequest struct {
	Token    string `json:"token" binding:"required,min=8,max=255"`
	Provider string `json:"provider"`
	Platform string `json:"platform"`
}

type UnregisterPushTokenRequest struct {
	Token string `json:"token" binding:"required,min=8,max=255"`
}

type PushTokenResponse struct {
	Token      string    `json:"token"`
	Provider   string    `json:"provider"`
	Platform   string    `json:"platform"`
	IsActive   bool      `json:"is_active"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

func ToNotificationResponse(notification *model.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        notification.ID.String(),
		UserID:    notification.UserID.String(),
		Type:      notification.Type,
		Title:     notification.Title,
		Body:      notification.Body,
		Data:      decodeMap(notification.Data),
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

func ToPushTokenResponse(pushToken *model.PushToken) PushTokenResponse {
	return PushTokenResponse{
		Token:      pushToken.Token,
		Provider:   pushToken.Provider,
		Platform:   pushToken.Platform,
		IsActive:   pushToken.IsActive,
		LastSeenAt: pushToken.LastSeenAt,
	}
}
