package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/metrics"
)

const (
	templateKeyPrefix     = "template:"
	templateTypeKeyPrefix = "template:type:"
)

// TemplateRepository implements repository.TemplateRepository using Redis
type TemplateRepository struct {
	client *redis.Client
}

// NewTemplateRepository creates a new Redis-based template repository
func NewTemplateRepository(client *redis.Client) *TemplateRepository {
	return &TemplateRepository{
		client: client,
	}
}

// Save saves a template to Redis
func (r *TemplateRepository) Save(ctx context.Context, template *model.Template) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_save_template", status, duration)
	}()

	// Marshal template to JSON
	data, err := json.Marshal(template)
	if err != nil {
		metrics.RecordOperationDuration("redis_save_template", "error", time.Since(start).Seconds())
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Create a transaction
	pipe := r.client.Pipeline()

	// Save template data
	key := fmt.Sprintf("%s%s", templateKeyPrefix, template.ID.String())
	pipe.Set(ctx, key, data, 0)

	// Add to type index
	typeKey := fmt.Sprintf("%s%s", templateTypeKeyPrefix, template.Type)
	pipe.SAdd(ctx, typeKey, template.ID.String())

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		metrics.RecordOperationDuration("redis_save_template", "error", time.Since(start).Seconds())
		return fmt.Errorf("failed to save template: %w", err)
	}

	return nil
}

// FindByID finds a template by ID from Redis
func (r *TemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_find_template_by_id", status, duration)
	}()

	key := fmt.Sprintf("%s%s", templateKeyPrefix, id.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		metrics.RecordOperationDuration("redis_find_template_by_id", "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var template model.Template
	if err := json.Unmarshal(data, &template); err != nil {
		metrics.RecordOperationDuration("redis_find_template_by_id", "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// FindByType finds templates by type from Redis
func (r *TemplateRepository) FindByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_find_templates_by_type", status, duration)
	}()

	typeKey := fmt.Sprintf("%s%s", templateTypeKeyPrefix, templateType)
	templateIDs, err := r.client.SMembers(ctx, typeKey).Result()
	if err != nil {
		metrics.RecordOperationDuration("redis_find_templates_by_type", "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("failed to get template IDs: %w", err)
	}

	templates := make([]*model.Template, 0, len(templateIDs))
	for _, id := range templateIDs {
		template, err := r.FindByID(ctx, uuid.MustParse(id))
		if err != nil {
			continue
		}
		if template != nil {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

// FindActiveByType finds active templates by type from Redis
func (r *TemplateRepository) FindActiveByType(ctx context.Context, templateType model.TemplateType) ([]*model.Template, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_find_active_templates_by_type", status, duration)
	}()

	templates, err := r.FindByType(ctx, templateType)
	if err != nil {
		return nil, err
	}

	activeTemplates := make([]*model.Template, 0, len(templates))
	for _, template := range templates {
		if template.IsActive {
			activeTemplates = append(activeTemplates, template)
		}
	}

	return activeTemplates, nil
}

// Update updates a template in Redis
func (r *TemplateRepository) Update(ctx context.Context, template *model.Template) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_update_template", status, duration)
	}()

	// Increment version
	template.Version++
	template.UpdatedAt = time.Now()

	// Save updated template
	err = r.Save(ctx, template)
	return err
}

// Delete deletes a template from Redis
func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("redis_delete_template", status, duration)
	}()

	// Get template to remove from type index
	template, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if template == nil {
		return nil
	}

	// Create a transaction
	pipe := r.client.Pipeline()

	// Delete template data
	key := fmt.Sprintf("%s%s", templateKeyPrefix, id.String())
	pipe.Del(ctx, key)

	// Remove from type index
	typeKey := fmt.Sprintf("%s%s", templateTypeKeyPrefix, template.Type)
	pipe.SRem(ctx, typeKey, id.String())

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		metrics.RecordOperationDuration("redis_delete_template", "error", time.Since(start).Seconds())
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}
