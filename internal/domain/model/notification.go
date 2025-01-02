package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	EmailNotification NotificationType = "email"
	SMSNotification   NotificationType = "sms"
	PushNotification  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
)

// Priority represents the priority level of a notification
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Notification represents a notification entity
type Notification struct {
	ID           uuid.UUID          `json:"id"`
	Type         NotificationType   `json:"type"`
	Status       NotificationStatus `json:"status"`
	Priority     Priority          `json:"priority"`
	Recipient    string            `json:"recipient"`
	Subject      string            `json:"subject"`
	Content      string            `json:"content"`
	TemplateID   string            `json:"template_id,omitempty"`
	TemplateData map[string]string `json:"template_data,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	SentAt       *time.Time        `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time        `json:"delivered_at,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// NewNotification creates a new notification
func NewNotification(
	notificationType NotificationType,
	recipient string,
	subject string,
	content string,
	priority Priority,
	templateID string,
	templateData map[string]string,
	metadata map[string]string,
) *Notification {
	now := time.Now()
	return &Notification{
		ID:           uuid.New(),
		Type:         notificationType,
		Status:       StatusPending,
		Priority:     priority,
		Recipient:    recipient,
		Subject:      subject,
		Content:      content,
		TemplateID:   templateID,
		TemplateData: templateData,
		Metadata:     metadata,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// MarkAsSent marks the notification as sent
func (n *Notification) MarkAsSent() {
	now := time.Now()
	n.Status = StatusSent
	n.SentAt = &now
	n.UpdatedAt = now
}

// MarkAsDelivered marks the notification as delivered
func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.Status = StatusDelivered
	n.DeliveredAt = &now
	n.UpdatedAt = now
}

// MarkAsFailed marks the notification as failed
func (n *Notification) MarkAsFailed(err error) {
	n.Status = StatusFailed
	n.Error = err.Error()
	n.UpdatedAt = time.Now()
}
