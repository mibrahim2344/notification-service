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

// NotificationRepository implements repository.NotificationRepository using PostgreSQL
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new PostgreSQL-based notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// Save saves a notification to PostgreSQL
func (r *NotificationRepository) Save(ctx context.Context, notification *model.Notification) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_save_notification", status, duration)
	}()

	templateData, err := json.Marshal(notification.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}

	metadata, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO notifications (
			id, recipient, type, subject, content, status, priority,
			template_id, template_type, template_data, metadata,
			error_message, retry_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err = r.db.ExecContext(ctx, query,
		notification.ID,
		notification.Recipient,
		notification.Type,
		notification.Subject,
		notification.Content,
		notification.Status,
		notification.Priority,
		notification.TemplateID,
		notification.TemplateType,
		templateData,
		metadata,
		notification.ErrorMessage,
		notification.RetryCount,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}

// FindByID finds a notification by ID from PostgreSQL
func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*model.Notification, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_notification_by_id", status, duration)
	}()

	// Convert string ID to UUID
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid notification ID format: %w", err)
	}

	query := `
		SELECT id, recipient, type, subject, content, status, priority,
			   template_id, template_type, template_data, metadata,
			   error_message, retry_count, created_at, updated_at
		FROM notifications
		WHERE id = $1`

	var notification model.Notification
	var templateData, metadata []byte

	err = r.db.QueryRowContext(ctx, query, uid).Scan(
		&notification.ID,
		&notification.Recipient,
		&notification.Type,
		&notification.Subject,
		&notification.Content,
		&notification.Status,
		&notification.Priority,
		&notification.TemplateID,
		&notification.TemplateType,
		&templateData,
		&metadata,
		&notification.ErrorMessage,
		&notification.RetryCount,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	if err := json.Unmarshal(templateData, &notification.TemplateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template data: %w", err)
	}

	if err := json.Unmarshal(metadata, &notification.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &notification, nil
}

// FindByRecipient finds notifications by recipient from PostgreSQL with pagination
func (r *NotificationRepository) FindByRecipient(ctx context.Context, recipient string, limit, offset int) ([]*model.Notification, error) {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_find_notifications_by_recipient", status, duration)
	}()

	query := `
		SELECT id, recipient, type, subject, content, status, priority,
			   template_id, template_type, template_data, metadata,
			   error_message, retry_count, created_at, updated_at
		FROM notifications
		WHERE recipient = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, recipient, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var notification model.Notification
		var templateData, metadata []byte

		err := rows.Scan(
			&notification.ID,
			&notification.Recipient,
			&notification.Type,
			&notification.Subject,
			&notification.Content,
			&notification.Status,
			&notification.Priority,
			&notification.TemplateID,
			&notification.TemplateType,
			&templateData,
			&metadata,
			&notification.ErrorMessage,
			&notification.RetryCount,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if err := json.Unmarshal(templateData, &notification.TemplateData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template data: %w", err)
		}

		if err := json.Unmarshal(metadata, &notification.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		notifications = append(notifications, &notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// Update updates a notification in PostgreSQL
func (r *NotificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_update_notification", status, duration)
	}()

	templateData, err := json.Marshal(notification.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}

	metadata, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE notifications
		SET recipient = $2,
			type = $3,
			subject = $4,
			content = $5,
			status = $6,
			priority = $7,
			template_id = $8,
			template_type = $9,
			template_data = $10,
			metadata = $11,
			error_message = $12,
			retry_count = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.Recipient,
		notification.Type,
		notification.Subject,
		notification.Content,
		notification.Status,
		notification.Priority,
		notification.TemplateID,
		notification.TemplateType,
		templateData,
		metadata,
		notification.ErrorMessage,
		notification.RetryCount,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", notification.ID)
	}

	return nil
}

// Delete deletes a notification from PostgreSQL
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	var err error
	defer func() {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		metrics.RecordOperationDuration("postgres_delete_notification", status, duration)
	}()

	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	return nil
}
