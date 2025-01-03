package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/mibrahim2344/notification-service/internal/domain/services"
	"go.uber.org/zap"
)

// Service implements the NotificationService interface
type Service struct {
	repo           services.NotificationRepository
	emailProvider  services.EmailProvider
	smsProvider    services.SMSProvider
	pushProvider   services.PushProvider
	templateEngine services.TemplateEngine
	logger         *zap.Logger
}

// NewService creates a new notification service
func NewService(
	repo services.NotificationRepository,
	emailProvider services.EmailProvider,
	smsProvider services.SMSProvider,
	pushProvider services.PushProvider,
	templateEngine services.TemplateEngine,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:           repo,
		emailProvider:  emailProvider,
		smsProvider:    smsProvider,
		pushProvider:   pushProvider,
		templateEngine: templateEngine,
		logger:         logger,
	}
}

// HandleUserEvent processes user-related events and sends appropriate notifications
func (s *Service) HandleUserEvent(ctx context.Context, eventType string, payload []byte) error {
	s.logger.Info("handling user event", zap.String("eventType", eventType))

	switch eventType {
	case "user.registered":
		return s.handleUserRegistered(ctx, payload)
	case "user.verified":
		return s.handleUserVerified(ctx, payload)
	case "user.password.reset":
		return s.handlePasswordReset(ctx, payload)
	case "user.password.changed":
		return s.handlePasswordChanged(ctx, payload)
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

func (s *Service) handleUserRegistered(ctx context.Context, payload []byte) error {
	var event struct {
		UserID    string `json:"userId"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("error unmarshaling user registered event: %w", err)
	}

	// Process welcome email template
	data := map[string]interface{}{
		"FirstName": event.FirstName,
		"Username":  event.Username,
		"Email":     event.Email,
		"Year":      time.Now().Year(),
	}

	content, err := s.templateEngine.ProcessTemplate(ctx, "welcome.html", data)
	if err != nil {
		return fmt.Errorf("error processing welcome template: %w", err)
	}

	notification := model.NewNotification(
		event.Email,
		model.EmailNotification,
		model.EmailTemplate,
		uuid.Nil,
		map[string]string{
			"subject":   "Welcome to Our Service",
			"content":   content,
			"eventType": "user.registered",
			"userId":    event.UserID,
		},
	)

	if err := s.repo.Save(ctx, notification); err != nil {
		return fmt.Errorf("error saving notification: %w", err)
	}

	if err := s.emailProvider.SendEmail(ctx, notification.Recipient, notification.Subject, notification.Content); err != nil {
		notification.UpdateStatus(model.StatusFailed, err.Error())
		if err := s.repo.Update(ctx, notification); err != nil {
			s.logger.Error("error updating notification status", zap.Error(err))
		}
		return fmt.Errorf("error sending welcome email: %w", err)
	}

	notification.UpdateStatus(model.StatusSent, "")
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error("error updating notification status", zap.Error(err))
	}

	return nil
}

func (s *Service) handleUserVerified(ctx context.Context, payload []byte) error {
	var event struct {
		UserID string `json:"userId"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("error unmarshaling user verified event: %w", err)
	}

	// Process verification success template
	data := map[string]interface{}{
		"Email": event.Email,
		"Year":  time.Now().Year(),
	}

	content, err := s.templateEngine.ProcessTemplate(ctx, "email_verified.html", data)
	if err != nil {
		return fmt.Errorf("error processing verification template: %w", err)
	}

	notification := model.NewNotification(
		event.Email,
		model.EmailNotification,
		model.EmailTemplate,
		uuid.Nil,
		map[string]string{
			"subject":   "Email Verification Successful",
			"content":   content,
			"eventType": "user.verified",
			"userId":    event.UserID,
		},
	)

	if err := s.repo.Save(ctx, notification); err != nil {
		return fmt.Errorf("error saving notification: %w", err)
	}

	if err := s.emailProvider.SendEmail(ctx, notification.Recipient, notification.Subject, notification.Content); err != nil {
		notification.UpdateStatus(model.StatusFailed, err.Error())
		if err := s.repo.Update(ctx, notification); err != nil {
			s.logger.Error("error updating notification status", zap.Error(err))
		}
		return fmt.Errorf("error sending verification email: %w", err)
	}

	notification.UpdateStatus(model.StatusSent, "")
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error("error updating notification status", zap.Error(err))
	}

	return nil
}

func (s *Service) handlePasswordReset(ctx context.Context, payload []byte) error {
	var event struct {
		UserID    string `json:"userId"`
		Email     string `json:"email"`
		ResetLink string `json:"resetLink"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("error unmarshaling password reset event: %w", err)
	}

	data := map[string]interface{}{
		"Email":     event.Email,
		"ResetLink": event.ResetLink,
		"Year":      time.Now().Year(),
	}

	content, err := s.templateEngine.ProcessTemplate(ctx, "password_reset.html", data)
	if err != nil {
		return fmt.Errorf("error processing password reset template: %w", err)
	}

	notification := model.NewNotification(
		event.Email,
		model.EmailNotification,
		model.EmailTemplate,
		uuid.Nil,
		map[string]string{
			"subject":   "Password Reset Request",
			"content":   content,
			"eventType": "user.password.reset",
			"userId":    event.UserID,
		},
	)

	if err := s.repo.Save(ctx, notification); err != nil {
		return fmt.Errorf("error saving notification: %w", err)
	}

	if err := s.emailProvider.SendEmail(ctx, notification.Recipient, notification.Subject, notification.Content); err != nil {
		notification.UpdateStatus(model.StatusFailed, err.Error())
		if err := s.repo.Update(ctx, notification); err != nil {
			s.logger.Error("error updating notification status", zap.Error(err))
		}
		return fmt.Errorf("error sending password reset email: %w", err)
	}

	notification.UpdateStatus(model.StatusSent, "")
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error("error updating notification status", zap.Error(err))
	}

	return nil
}

func (s *Service) handlePasswordChanged(ctx context.Context, payload []byte) error {
	var event struct {
		UserID string `json:"userId"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("error unmarshaling password changed event: %w", err)
	}

	data := map[string]interface{}{
		"Email": event.Email,
		"Year":  time.Now().Year(),
	}

	content, err := s.templateEngine.ProcessTemplate(ctx, "password_changed.html", data)
	if err != nil {
		return fmt.Errorf("error processing password changed template: %w", err)
	}

	notification := model.NewNotification(
		event.Email,
		model.EmailNotification,
		model.EmailTemplate,
		uuid.Nil,
		map[string]string{
			"subject":   "Password Changed Successfully",
			"content":   content,
			"eventType": "user.password.changed",
			"userId":    event.UserID,
		},
	)

	if err := s.repo.Save(ctx, notification); err != nil {
		return fmt.Errorf("error saving notification: %w", err)
	}

	if err := s.emailProvider.SendEmail(ctx, notification.Recipient, notification.Subject, notification.Content); err != nil {
		notification.UpdateStatus(model.StatusFailed, err.Error())
		if err := s.repo.Update(ctx, notification); err != nil {
			s.logger.Error("error updating notification status", zap.Error(err))
		}
		return fmt.Errorf("error sending password changed email: %w", err)
	}

	notification.UpdateStatus(model.StatusSent, "")
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error("error updating notification status", zap.Error(err))
	}

	return nil
}

// Other interface methods implementation...
func (s *Service) SendNotification(ctx context.Context, notification *model.Notification) error {
	if err := s.repo.Save(ctx, notification); err != nil {
		return fmt.Errorf("error saving notification: %w", err)
	}

	var err error
	switch notification.Type {
	case model.EmailNotification:
		err = s.emailProvider.SendEmail(ctx, notification.Recipient, notification.Subject, notification.Content)
	case model.SMSNotification:
		err = s.smsProvider.SendSMS(ctx, notification.Recipient, notification.Content)
	case model.PushNotification:
		err = s.pushProvider.SendPush(ctx, notification.Recipient, notification.Subject, notification.Content)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	if err != nil {
		notification.UpdateStatus(model.StatusFailed, err.Error())
		if updateErr := s.repo.Update(ctx, notification); updateErr != nil {
			s.logger.Error("error updating notification status", zap.Error(updateErr))
		}
		return fmt.Errorf("error sending notification: %w", err)
	}

	notification.UpdateStatus(model.StatusSent, "")
	if err := s.repo.Update(ctx, notification); err != nil {
		s.logger.Error("error updating notification status", zap.Error(err))
	}

	return nil
}

func (s *Service) GetNotification(ctx context.Context, id string) (*model.Notification, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetNotificationHistory(ctx context.Context, recipient string, limit, offset int) ([]*model.Notification, error) {
	return s.repo.FindByRecipient(ctx, recipient, limit, offset)
}

func (s *Service) GetNotificationsByRecipient(recipient string, limit, offset int) ([]*model.Notification, error) {
	return s.GetNotificationHistory(context.Background(), recipient, limit, offset)
}
