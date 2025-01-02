package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(notification *model.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotificationService) GetNotification(id string) (*model.Notification, error) {
	args := m.Called(id)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	if args.Get(0) == nil {
		return nil, nil
	}
	return args.Get(0).(*model.Notification), nil
}

func (m *MockNotificationService) GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error) {
	args := m.Called(recipient, limit, offset)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), nil
}

func TestNotificationHandler_SendNotification(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockNotificationService)
	handler := NewNotificationHandler(mockService, logger)

	tests := []struct {
		name           string
		request        SendNotificationRequest
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "successful notification send",
			request: SendNotificationRequest{
				Recipient: "test@example.com",
				Type:     "email",
				Subject:  "Test Subject",
				Content:  "Test Content",
				Priority: "high",
			},
			setupMock: func() {
				mockService.On("SendNotification", mock.AnythingOfType("*model.Notification")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "service error",
			request: SendNotificationRequest{
				Recipient: "test@example.com",
				Type:     "email",
				Subject:  "Test Subject",
				Content:  "Test Content",
				Priority: "high",
			},
			setupMock: func() {
				mockService.On("SendNotification", mock.AnythingOfType("*model.Notification")).Return(assert.AnError)
			},
			expectedStatus: http.StatusFailedDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Setup
			tt.setupMock()

			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			// Execute request
			handler.SendNotification(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_GetNotification(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockNotificationService)
	handler := NewNotificationHandler(mockService, logger)

	notification := &model.Notification{
		ID:        uuid.New(),
		Recipient: "test@example.com",
		Type:      model.EmailNotification,
		Subject:   "Test Subject",
		Content:   "Test Content",
		Status:    model.StatusSent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		notificationID string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:           "successful get",
			notificationID: notification.ID.String(),
			setupMock: func() {
				mockService.On("GetNotification", notification.ID.String()).Return(notification, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			notificationID: "non-existent",
			setupMock: func() {
				mockService.On("GetNotification", "non-existent").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "service error",
			notificationID: notification.ID.String(),
			setupMock: func() {
				mockService.On("GetNotification", notification.ID.String()).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusFailedDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Setup
			tt.setupMock()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/notifications/"+tt.notificationID, nil)
			rec := httptest.NewRecorder()

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.notificationID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute request
			handler.GetNotification(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestNotificationHandler_GetNotificationsByRecipient(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockNotificationService)
	handler := NewNotificationHandler(mockService, logger)

	notifications := []*model.Notification{
		{
			ID:        uuid.New(),
			Recipient: "test@example.com",
			Type:      model.EmailNotification,
			Subject:   "Test Subject 1",
			Content:   "Test Content 1",
			Status:    model.StatusSent,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Recipient: "test@example.com",
			Type:      model.EmailNotification,
			Subject:   "Test Subject 2",
			Content:   "Test Content 2",
			Status:    model.StatusSent,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tests := []struct {
		name           string
		recipient     string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "successful get",
			recipient: "test@example.com",
			setupMock: func() {
				mockService.On("GetNotificationsByRecipient", "test@example.com", 10, 0).Return(notifications, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "missing recipient",
			recipient: "",
			setupMock: func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			recipient: "test@example.com",
			setupMock: func() {
				mockService.On("GetNotificationsByRecipient", "test@example.com", 10, 0).Return([]*model.Notification(nil), assert.AnError)
			},
			expectedStatus: http.StatusFailedDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil

			// Setup
			tt.setupMock()

			// Create request
			url := "/notifications"
			if tt.recipient != "" {
				url += "?recipient=" + tt.recipient
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			// Execute request
			handler.GetNotificationsByRecipient(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockService.AssertExpectations(t)
		})
	}
}
