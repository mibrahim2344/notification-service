package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Notification types
	EmailNotification NotificationType = "email"
	SMSNotification  NotificationType = "sms"
	PushNotification NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	// Notification statuses
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

// Priority represents the priority level of a notification
type Priority string

const (
	// Priority levels
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// TemplateType represents the type of template
type TemplateType string

const (
	// Template types
	EmailTemplate TemplateType = "email"
	SMSTemplate  TemplateType = "sms"
	PushTemplate TemplateType = "push"
)

// Notification represents a notification entity
type Notification struct {
	ID           uuid.UUID          `json:"id" redis:"id"`
	Recipient    string            `json:"recipient" redis:"recipient"`
	Type         NotificationType  `json:"type" redis:"type"`
	Subject      string            `json:"subject" redis:"subject"`
	Content      string            `json:"content" redis:"content"`
	Status       NotificationStatus `json:"status" redis:"status"`
	Priority     Priority          `json:"priority" redis:"priority"`
	TemplateID   uuid.UUID         `json:"template_id,omitempty" redis:"template_id"`
	TemplateType TemplateType      `json:"template_type,omitempty" redis:"template_type"`
	TemplateData map[string]string `json:"template_data,omitempty" redis:"template_data"`
	Metadata     map[string]string `json:"metadata,omitempty" redis:"metadata"`
	ErrorMessage string            `json:"error_message,omitempty" redis:"error_message"`
	RetryCount   int               `json:"retry_count" redis:"retry_count"`
	CreatedAt    time.Time         `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" redis:"updated_at"`
}

// NewNotification creates a new notification
func NewNotification(recipient string, notificationType NotificationType, templateType TemplateType, templateID uuid.UUID, templateData map[string]string) *Notification {
	now := time.Now()
	return &Notification{
		ID:           uuid.New(),
		Recipient:    recipient,
		Type:         notificationType,
		Status:       StatusPending,
		Priority:     PriorityMedium,
		TemplateID:   templateID,
		TemplateType: templateType,
		TemplateData: templateData,
		RetryCount:   0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Validate validates the notification
func (n *Notification) Validate() error {
	if n.Recipient == "" {
		return ErrInvalidNotification{Message: "recipient is required"}
	}
	if n.Type == "" {
		return ErrInvalidNotification{Message: "notification type is required"}
	}
	if n.TemplateID == uuid.Nil {
		return ErrInvalidNotification{Message: "template ID is required"}
	}
	if n.TemplateType == "" {
		return ErrInvalidNotification{Message: "template type is required"}
	}
	return nil
}

// UpdateStatus updates the notification status
func (n *Notification) UpdateStatus(status NotificationStatus, errorMessage string) {
	n.Status = status
	n.ErrorMessage = errorMessage
	n.UpdatedAt = time.Now()
}

// IncrementRetryCount increments the retry count
func (n *Notification) IncrementRetryCount() {
	n.RetryCount++
	n.UpdatedAt = time.Now()
}

// ErrInvalidNotification represents a notification validation error
type ErrInvalidNotification struct {
	Message string
}

func (e ErrInvalidNotification) Error() string {
	return e.Message
}
