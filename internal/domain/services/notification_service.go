package services

import (
	"context"

	"github.com/mibrahim2344/notification-service/internal/domain/model"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// SendNotification sends a notification
	SendNotification(ctx context.Context, notification *model.Notification) error

	// GetNotification retrieves a notification by ID
	GetNotification(ctx context.Context, id string) (*model.Notification, error)

	// GetNotificationHistory retrieves notification history for a recipient
	GetNotificationHistory(ctx context.Context, recipient string, limit, offset int) ([]*model.Notification, error)

	// HandleUserEvent processes user-related events and sends appropriate notifications
	HandleUserEvent(ctx context.Context, eventType string, payload []byte) error
}

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, content string) error
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	SendSMS(ctx context.Context, to, message string) error
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	SendPush(ctx context.Context, token, title, message string) error
}

// TemplateEngine defines the interface for template processing
type TemplateEngine interface {
	// ProcessTemplate processes a template with given data
	ProcessTemplate(ctx context.Context, templateName string, data interface{}) (string, error)

	// GetTemplate retrieves a template by name and locale
	GetTemplate(ctx context.Context, templateName, locale string) (string, error)
}

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	Save(ctx context.Context, notification *model.Notification) error
	FindByID(ctx context.Context, id string) (*model.Notification, error)
	FindByRecipient(ctx context.Context, recipient string, limit, offset int) ([]*model.Notification, error)
	Update(ctx context.Context, notification *model.Notification) error
}
