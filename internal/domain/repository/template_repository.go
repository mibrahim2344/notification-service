package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
)

// TemplateRepository defines the interface for template storage operations
type TemplateRepository interface {
	// Save saves a template
	Save(ctx context.Context, template *model.Template) error

	// FindByID finds a template by ID
	FindByID(ctx context.Context, id uuid.UUID) (*model.Template, error)

	// FindByType finds templates by type
	FindByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error)

	// FindActiveByType finds active templates by type
	FindActiveByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error)

	// Update updates a template
	Update(ctx context.Context, template *model.Template) error

	// Delete deletes a template
	Delete(ctx context.Context, id uuid.UUID) error
}
