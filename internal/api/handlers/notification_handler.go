package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	notificationService NotificationService
	logger             *zap.Logger
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	SendNotification(notification *model.Notification) error
	GetNotification(id string) (*model.Notification, error)
	GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error)
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: service,
		logger:             logger,
	}
}

// SendNotificationRequest represents the request body for sending a notification
type SendNotificationRequest struct {
	Recipient    string            `json:"recipient" validate:"required,email"`
	Type         string            `json:"type" validate:"required,oneof=email sms push"`
	Subject      string            `json:"subject" validate:"required"`
	Content      string            `json:"content" validate:"required"`
	Priority     string            `json:"priority" validate:"required,oneof=high medium low"`
	TemplateID   string            `json:"template_id,omitempty"`
	TemplateData map[string]string `json:"template_data,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NotificationResponse represents the response for notification operations
type NotificationResponse struct {
	ID        string            `json:"id"`
	Recipient string            `json:"recipient"`
	Type      string            `json:"type"`
	Subject   string            `json:"subject"`
	Content   string            `json:"content"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// RegisterRoutes registers the notification routes
func (h *NotificationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/notifications", h.SendNotification)
	r.Get("/notifications/{id}", h.GetNotification)
	r.Get("/notifications", h.GetNotificationsByRecipient)
}

func writeError(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}

func writeResponse(w http.ResponseWriter, data interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

// SendNotification handles the notification sending request
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	operation := "send_notification"

	var req SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	notification := &model.Notification{
		ID:           uuid.New(),
		Recipient:    req.Recipient,
		Type:         model.NotificationType(req.Type),
		Subject:      req.Subject,
		Content:      req.Content,
		Priority:     model.Priority(req.Priority),
		Status:       model.StatusPending,
		TemplateID:   req.TemplateID,
		TemplateData: req.TemplateData,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.notificationService.SendNotification(notification); err != nil {
		h.logger.Error("failed to send notification",
			zap.Error(err),
			zap.String("recipient", req.Recipient),
			zap.String("type", req.Type),
		)
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to send notification", http.StatusFailedDependency)
		return
	}

	response := NotificationResponse{
		ID:        notification.ID.String(),
		Recipient: notification.Recipient,
		Type:      string(notification.Type),
		Subject:   notification.Subject,
		Content:   notification.Content,
		Status:    string(notification.Status),
		Metadata:  notification.Metadata,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}

	if err := writeResponse(w, response, http.StatusCreated); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	metrics.RecordOperationDuration("http_"+operation, "success", time.Since(start).Seconds())
}

// GetNotification handles the request to get a notification by ID
func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	operation := "get_notification"

	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.Error("notification ID is required")
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	notification, err := h.notificationService.GetNotification(id)
	if err != nil {
		h.logger.Error("failed to get notification",
			zap.Error(err),
			zap.String("id", id),
		)
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to get notification", http.StatusFailedDependency)
		return
	}

	if notification == nil {
		metrics.RecordOperationDuration("http_"+operation, "not_found", time.Since(start).Seconds())
		writeError(w, "Notification not found", http.StatusNotFound)
		return
	}

	response := NotificationResponse{
		ID:        notification.ID.String(),
		Recipient: notification.Recipient,
		Type:      string(notification.Type),
		Subject:   notification.Subject,
		Content:   notification.Content,
		Status:    string(notification.Status),
		Metadata:  notification.Metadata,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}

	if err := writeResponse(w, response, http.StatusOK); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	metrics.RecordOperationDuration("http_"+operation, "success", time.Since(start).Seconds())
}

// GetNotificationsByRecipient handles the request to get notifications for a recipient
func (h *NotificationHandler) GetNotificationsByRecipient(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	operation := "get_notifications_by_recipient"

	recipient := r.URL.Query().Get("recipient")
	if recipient == "" {
		h.logger.Error("recipient is required")
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Recipient is required", http.StatusBadRequest)
		return
	}

	limit := 10 // Default limit
	offset := 0 // Default offset

	notifications, err := h.notificationService.GetNotificationsByRecipient(recipient, limit, offset)
	if err != nil {
		h.logger.Error("failed to get notifications",
			zap.Error(err),
			zap.String("recipient", recipient),
		)
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to get notifications", http.StatusFailedDependency)
		return
	}

	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, NotificationResponse{
			ID:        notification.ID.String(),
			Recipient: notification.Recipient,
			Type:      string(notification.Type),
			Subject:   notification.Subject,
			Content:   notification.Content,
			Status:    string(notification.Status),
			Metadata:  notification.Metadata,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.UpdatedAt,
		})
	}

	if err := writeResponse(w, response, http.StatusOK); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		metrics.RecordOperationDuration("http_"+operation, "error", time.Since(start).Seconds())
		writeError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	metrics.RecordOperationDuration("http_"+operation, "success", time.Since(start).Seconds())
}
