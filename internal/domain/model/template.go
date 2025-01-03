package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Template types
	WelcomeEmail      TemplateType = "welcome_email"
	TwoFactorAuth     TemplateType = "2fa"
	PasswordReset     TemplateType = "password_reset"
	AccountActivation TemplateType = "account_activation"
)

// Template represents a notification template
type Template struct {
	ID        uuid.UUID         `json:"id" redis:"id"`
	Name      string            `json:"name" redis:"name"`
	Type      TemplateType      `json:"type" redis:"type"`
	Subject   string            `json:"subject" redis:"subject"`
	Content   string            `json:"content" redis:"content"`
	Variables []string          `json:"variables" redis:"variables"`
	Metadata  map[string]string `json:"metadata,omitempty" redis:"metadata"`
	Version   int               `json:"version" redis:"version"`
	IsActive  bool              `json:"is_active" redis:"is_active"`
	CreatedAt time.Time         `json:"created_at" redis:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" redis:"updated_at"`
}

// NewTemplate creates a new template
func NewTemplate(name string, templateType TemplateType, subject, content string) *Template {
	now := time.Now()
	return &Template{
		ID:        uuid.New(),
		Name:      name,
		Type:      templateType,
		Subject:   subject,
		Content:   content,
		Version:   1,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the template
func (t *Template) Validate() error {
	if t.Name == "" {
		return ErrInvalidTemplate{Message: "template name is required"}
	}
	if t.Type == "" {
		return ErrInvalidTemplate{Message: "template type is required"}
	}
	if t.Subject == "" {
		return ErrInvalidTemplate{Message: "template subject is required"}
	}
	if t.Content == "" {
		return ErrInvalidTemplate{Message: "template content is required"}
	}
	return nil
}

// ErrInvalidTemplate represents a template validation error
type ErrInvalidTemplate struct {
	Message string
}

func (e ErrInvalidTemplate) Error() string {
	return e.Message
}
