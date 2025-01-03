package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/metrics"
)

// TemplateRepository implements repository.TemplateRepository using PostgreSQL
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new PostgreSQL-based template repository
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{
		db: db,
	}
}

// Save saves a template to PostgreSQL
func (r *TemplateRepository) Save(ctx context.Context, template *model.Template) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_save_template", status, duration)
	}()

	variables, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	metadata, err := json.Marshal(template.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO templates (
			id, name, type, subject, content, variables, metadata,
			version, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err = r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Type,
		template.Subject,
		template.Content,
		variables,
		metadata,
		template.Version,
		template.IsActive,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID finds a template by ID from PostgreSQL
func (r *TemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_template_by_id", status, duration)
	}()

	query := `
		SELECT id, name, type, subject, content, variables, metadata,
			   version, is_active, created_at, updated_at
		FROM templates
		WHERE id = $1`

	var template model.Template
	var variables, metadata []byte

	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Type,
		&template.Subject,
		&template.Content,
		&variables,
		&metadata,
		&template.Version,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	if err := json.Unmarshal(variables, &template.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	if err := json.Unmarshal(metadata, &template.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &template, nil
}

// FindByType finds templates by type from PostgreSQL
func (r *TemplateRepository) FindByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_templates_by_type", status, duration)
	}()

	query := `
		SELECT id, name, type, subject, content, variables, metadata,
			   version, is_active, created_at, updated_at
		FROM templates
		WHERE type = $1
		ORDER BY version DESC`

	rows, err := r.db.QueryContext(ctx, query, templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*model.Template
	for rows.Next() {
		var template model.Template
		var variables, metadata []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Type,
			&template.Subject,
			&template.Content,
			&variables,
			&metadata,
			&template.Version,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if err := json.Unmarshal(variables, &template.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		if err := json.Unmarshal(metadata, &template.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// FindActiveByType finds active templates by type from PostgreSQL
func (r *TemplateRepository) FindActiveByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_active_templates_by_type", status, duration)
	}()

	query := `
		SELECT id, name, type, subject, content, variables, metadata,
			   version, is_active, created_at, updated_at
		FROM templates
		WHERE type = $1 AND is_active = true
		ORDER BY version DESC`

	rows, err := r.db.QueryContext(ctx, query, templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*model.Template
	for rows.Next() {
		var template model.Template
		var variables, metadata []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Type,
			&template.Subject,
			&template.Content,
			&variables,
			&metadata,
			&template.Version,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		if err := json.Unmarshal(variables, &template.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		if err := json.Unmarshal(metadata, &template.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// Update updates a template in PostgreSQL
func (r *TemplateRepository) Update(ctx context.Context, template *model.Template) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_update_template", status, duration)
	}()

	variables, err := json.Marshal(template.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	metadata, err := json.Marshal(template.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE templates
		SET name = $2,
			type = $3,
			subject = $4,
			content = $5,
			variables = $6,
			metadata = $7,
			version = $8,
			is_active = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Type,
		template.Subject,
		template.Content,
		variables,
		metadata,
		template.Version,
		template.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	return nil
}

// Delete deletes a template from PostgreSQL
func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_delete_template", status, duration)
	}()

	query := `DELETE FROM templates WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", id)
	}

	return nil
}

// ProcessTemplate processes a template with given data
func (r *TemplateRepository) ProcessTemplate(ctx context.Context, templateName string, data interface{}) (string, error) {
	// Find the template by name
	template, err := r.findByName(ctx, templateName)
	if err != nil {
		return "", fmt.Errorf("failed to find template: %w", err)
	}

	// TODO: Implement actual template processing logic
	// For now, return the raw content
	return template.Content, nil
}

// GetTemplate retrieves a template by name and locale
func (r *TemplateRepository) GetTemplate(ctx context.Context, templateName, locale string) (string, error) {
	template, err := r.findByName(ctx, templateName)
	if err != nil {
		return "", fmt.Errorf("failed to find template: %w", err)
	}

	return template.Content, nil
}

// findByName finds a template by name from PostgreSQL
func (r *TemplateRepository) findByName(ctx context.Context, name string) (*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_template_by_name", status, duration)
	}()

	query := `
		SELECT id, name, type, content, variables, metadata, is_active, created_at, updated_at
		FROM templates
		WHERE name = $1 AND is_active = true
		LIMIT 1`

	var template model.Template
	var variables, metadata []byte

	err = r.db.QueryRowContext(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Type,
		&template.Content,
		&variables,
		&metadata,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	if err := json.Unmarshal(variables, &template.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	if err := json.Unmarshal(metadata, &template.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &template, nil
}
