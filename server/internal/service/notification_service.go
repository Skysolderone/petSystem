package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"petverse/server/internal/dto"
	"petverse/server/internal/model"
	"petverse/server/internal/pkg/apperror"
	"petverse/server/internal/pkg/events"
	"petverse/server/internal/pkg/pagination"
	"petverse/server/internal/pkg/push"
	"petverse/server/internal/ws"
)

type NotificationCreateInput struct {
	Type  string
	Title string
	Body  string
	Data  map[string]any
}

type NotificationService struct {
	notifications notificationRepository
	pushTokens    pushTokenRepository
	dispatcher    push.Dispatcher
	events        events.Publisher
	hub           *ws.Hub
}

type notificationRepository interface {
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUnreadByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	Create(ctx context.Context, notification *model.Notification) error
	CreateBatch(ctx context.Context, notifications []model.Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Notification, int64, error)
	LatestByUser(ctx context.Context, userID uuid.UUID, limit int) ([]model.Notification, error)
	GetByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*model.Notification, error)
	Update(ctx context.Context, notification *model.Notification) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error)
}

type pushTokenRepository interface {
	Upsert(ctx context.Context, pushToken *model.PushToken) error
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]model.PushToken, error)
	DeactivateByUserAndToken(ctx context.Context, userID uuid.UUID, token string) error
	DeactivateTokens(ctx context.Context, tokens []string) error
}

type NotificationServiceOption func(*NotificationService)

func WithPushNotifications(repo pushTokenRepository, dispatcher push.Dispatcher) NotificationServiceOption {
	return func(service *NotificationService) {
		service.pushTokens = repo
		service.dispatcher = dispatcher
	}
}

func WithNotificationEvents(publisher events.Publisher) NotificationServiceOption {
	return func(service *NotificationService) {
		service.events = publisher
	}
}

func NewNotificationService(notifications notificationRepository, hub *ws.Hub, options ...NotificationServiceOption) *NotificationService {
	service := &NotificationService{
		notifications: notifications,
		hub:           hub,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Notification, int64, int, int, error) {
	if err := s.ensureSeed(ctx, userID); err != nil {
		return nil, 0, 0, 0, err
	}

	page, pageSize = pagination.Normalize(page, pageSize)
	items, total, err := s.notifications.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, apperror.Wrap(http.StatusInternalServerError, "list_notifications_failed", "failed to load notifications", err)
	}
	return items, total, page, pageSize, nil
}

func (s *NotificationService) Snapshot(ctx context.Context, userID uuid.UUID) ([]model.Notification, int64, error) {
	if err := s.ensureSeed(ctx, userID); err != nil {
		return nil, 0, err
	}

	items, err := s.notifications.LatestByUser(ctx, userID, 10)
	if err != nil {
		return nil, 0, apperror.Wrap(http.StatusInternalServerError, "list_notifications_failed", "failed to load notifications", err)
	}
	unreadCount, err := s.notifications.CountUnreadByUser(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(http.StatusInternalServerError, "count_notifications_failed", "failed to count notifications", err)
	}
	return items, unreadCount, nil
}

func (s *NotificationService) Create(ctx context.Context, userID uuid.UUID, input NotificationCreateInput) (*model.Notification, error) {
	notificationType := input.Type
	if notificationType == "" {
		notificationType = "system"
	}

	notification := &model.Notification{
		UserID: userID,
		Type:   notificationType,
		Title:  input.Title,
		Body:   input.Body,
		Data:   datatypes.JSON(dto.EncodeMap(input.Data)),
	}
	if err := s.notifications.Create(ctx, notification); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "create_notification_failed", "failed to create notification", err)
	}

	if s.hub != nil {
		unreadCount, err := s.notifications.CountUnreadByUser(ctx, userID)
		if err == nil {
			resolved := dto.ToNotificationResponse(notification)
			s.hub.BroadcastUser(userID, dto.NotificationStreamMessage{
				Type:         "notification",
				Timestamp:    time.Now(),
				Notification: &resolved,
				UnreadCount:  unreadCount,
			})
		}
	}

	s.dispatchPush(ctx, userID, notification)
	s.publish(ctx, "notification.created", map[string]any{
		"notification_id": notification.ID.String(),
		"user_id":         userID.String(),
		"type":            notification.Type,
		"title":           notification.Title,
	})

	return notification, nil
}

func (s *NotificationService) RegisterPushToken(ctx context.Context, userID uuid.UUID, req dto.RegisterPushTokenRequest) (*model.PushToken, error) {
	if s.pushTokens == nil {
		return nil, apperror.New(http.StatusNotImplemented, "push_not_enabled", "push notifications are not enabled")
	}

	now := time.Now()
	pushToken := &model.PushToken{
		UserID:     userID,
		Provider:   firstNonEmpty(req.Provider, "expo"),
		Token:      req.Token,
		Platform:   firstNonEmpty(req.Platform, "unknown"),
		IsActive:   true,
		LastSeenAt: now,
	}
	if err := s.pushTokens.Upsert(ctx, pushToken); err != nil {
		return nil, apperror.Wrap(http.StatusInternalServerError, "register_push_token_failed", "failed to register push token", err)
	}
	return pushToken, nil
}

func (s *NotificationService) UnregisterPushToken(ctx context.Context, userID uuid.UUID, token string) error {
	if s.pushTokens == nil {
		return nil
	}
	if err := s.pushTokens.DeactivateByUserAndToken(ctx, userID, token); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "unregister_push_token_failed", "failed to unregister push token", err)
	}
	return nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) (*model.Notification, int64, error) {
	notification, err := s.notifications.GetByIDAndUser(ctx, notificationID, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(http.StatusInternalServerError, "get_notification_failed", "failed to load notification", err)
	}
	if notification == nil {
		return nil, 0, apperror.New(http.StatusNotFound, "notification_not_found", "notification not found")
	}

	if !notification.IsRead {
		notification.IsRead = true
		if err := s.notifications.Update(ctx, notification); err != nil {
			return nil, 0, apperror.Wrap(http.StatusInternalServerError, "update_notification_failed", "failed to update notification", err)
		}
	}

	unreadCount, err := s.notifications.CountUnreadByUser(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(http.StatusInternalServerError, "count_notifications_failed", "failed to count notifications", err)
	}
	return notification, unreadCount, nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, int64, error) {
	updated, err := s.notifications.MarkAllRead(ctx, userID)
	if err != nil {
		return 0, 0, apperror.Wrap(http.StatusInternalServerError, "update_notifications_failed", "failed to update notifications", err)
	}
	unreadCount, err := s.notifications.CountUnreadByUser(ctx, userID)
	if err != nil {
		return 0, 0, apperror.Wrap(http.StatusInternalServerError, "count_notifications_failed", "failed to count notifications", err)
	}
	return updated, unreadCount, nil
}

func (s *NotificationService) ensureSeed(ctx context.Context, userID uuid.UUID) error {
	count, err := s.notifications.CountByUser(ctx, userID)
	if err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "count_notifications_failed", "failed to count notifications", err)
	}
	if count > 0 {
		return nil
	}

	seedNotifications := []model.Notification{
		{
			UserID: userID,
			Type:   "system",
			Title:  "欢迎来到 PetVerse",
			Body:   "你的宠物健康、设备和服务预约现在已经可以统一管理。",
			Data:   datatypes.JSON(dto.EncodeMap(map[string]any{"route": "/(tabs)"})),
		},
		{
			UserID: userID,
			Type:   "promotion",
			Title:  "Phase 4 能力已开放",
			Body:   "可以试试 AI 训练计划和智能商品推荐。",
			Data:   datatypes.JSON(dto.EncodeMap(map[string]any{"route": "/training"})),
		},
	}
	if err := s.notifications.CreateBatch(ctx, seedNotifications); err != nil {
		return apperror.Wrap(http.StatusInternalServerError, "seed_notifications_failed", "failed to seed notifications", err)
	}
	return nil
}

func (s *NotificationService) dispatchPush(ctx context.Context, userID uuid.UUID, notification *model.Notification) {
	if s.pushTokens == nil || s.dispatcher == nil || notification == nil {
		return
	}

	items, err := s.pushTokens.ListActiveByUser(ctx, userID)
	if err != nil || len(items) == 0 {
		return
	}

	tokens := make([]string, 0, len(items))
	for _, item := range items {
		if item.Token != "" {
			tokens = append(tokens, item.Token)
		}
	}
	invalidTokens, err := s.dispatcher.Send(ctx, tokens, push.Payload{
		Title: notification.Title,
		Body:  notification.Body,
		Data:  dto.DecodeJSONMap(notification.Data),
	})
	if err == nil && len(invalidTokens) > 0 {
		_ = s.pushTokens.DeactivateTokens(ctx, invalidTokens)
	}
}

func (s *NotificationService) publish(ctx context.Context, subject string, payload any) {
	if s.events == nil {
		return
	}
	_ = s.events.PublishJSON(ctx, subject, payload)
}
