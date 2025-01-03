package services

import (
	"context"

	"github.com/mibrahim2344/notification-service/internal/domain/model"
)

// NotificationServiceAdapter adapts the domain notification service to the handler interface
type NotificationServiceAdapter struct {
	service interface {
		SendNotification(ctx context.Context, notification *model.Notification) error
		GetNotification(ctx context.Context, id string) (*model.Notification, error)
		GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error)
	}
}

// NewNotificationServiceAdapter creates a new notification service adapter
func NewNotificationServiceAdapter(service interface {
	SendNotification(ctx context.Context, notification *model.Notification) error
	GetNotification(ctx context.Context, id string) (*model.Notification, error)
	GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error)
}) *NotificationServiceAdapter {
	return &NotificationServiceAdapter{
		service: service,
	}
}

// SendNotification adapts the domain service's SendNotification method to the handler interface
func (a *NotificationServiceAdapter) SendNotification(notification *model.Notification) error {
	return a.service.SendNotification(context.Background(), notification)
}

// GetNotification adapts the domain service's GetNotification method to the handler interface
func (a *NotificationServiceAdapter) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	return a.service.GetNotification(ctx, id)
}

// GetNotificationsByRecipient adapts the domain service's GetNotificationsByRecipient method to the handler interface
func (a *NotificationServiceAdapter) GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error) {
	return a.service.GetNotificationsByRecipient(recipient, limit, offset)
}
