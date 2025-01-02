package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/metrics"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Key prefixes
	notificationPrefix = "notification:"
	recipientPrefix   = "recipient:"
	
	// Default expiration for notifications (30 days)
	defaultExpiration = 30 * 24 * time.Hour
)

// NotificationRepository implements repository interface using Redis
type NotificationRepository struct {
	client *redis.Client
	logger *zap.Logger
}

// NewNotificationRepository creates a new Redis-based notification repository
func NewNotificationRepository(client *redis.Client, logger *zap.Logger) *NotificationRepository {
	// Set initial connection status
	metrics.SetRedisConnectionStatus(true)

	return &NotificationRepository{
		client: client,
		logger: logger,
	}
}

// Save stores a notification in Redis
func (r *NotificationRepository) Save(ctx context.Context, notification *model.Notification) error {
	start := time.Now()
	operation := "save"

	// Marshal notification to JSON
	data, err := json.Marshal(notification)
	if err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	// Update storage size metric
	metrics.UpdateNotificationStorageSize(string(notification.Type), float64(len(data)))

	// Create pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store notification data
	notificationKey := fmt.Sprintf("%s%s", notificationPrefix, notification.ID)
	pipe.Set(ctx, notificationKey, data, defaultExpiration)

	// Add to recipient's notification list
	recipientKey := fmt.Sprintf("%s%s", recipientPrefix, notification.Recipient)
	pipe.ZAdd(ctx, recipientKey, redis.Z{
		Score:  float64(notification.CreatedAt.Unix()),
		Member: notification.ID.String(),
	})
	pipe.Expire(ctx, recipientKey, defaultExpiration)

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error saving notification: %w", err)
	}

	metrics.RecordOperationDuration(operation, "success", time.Since(start).Seconds())
	metrics.UpdateNotificationStatus(string(notification.Status), 1)
	return nil
}

// FindByID retrieves a notification by ID
func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*model.Notification, error) {
	start := time.Now()
	operation := "find_by_id"

	key := fmt.Sprintf("%s%s", notificationPrefix, id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			metrics.RecordCacheMiss()
			metrics.RecordOperationDuration(operation, "not_found", time.Since(start).Seconds())
			return nil, nil // Not found
		}
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("error retrieving notification: %w", err)
	}

	metrics.RecordCacheHit()

	var notification model.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("error unmarshaling notification: %w", err)
	}

	metrics.RecordOperationDuration(operation, "success", time.Since(start).Seconds())
	return &notification, nil
}

// FindByRecipient retrieves notifications for a recipient with pagination
func (r *NotificationRepository) FindByRecipient(ctx context.Context, recipient string, limit, offset int) ([]*model.Notification, error) {
	start := time.Now()
	operation := "find_by_recipient"

	// Get notification IDs from sorted set
	recipientKey := fmt.Sprintf("%s%s", recipientPrefix, recipient)
	ids, err := r.client.ZRevRange(ctx, recipientKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("error retrieving notification IDs: %w", err)
	}

	if len(ids) == 0 {
		metrics.RecordOperationDuration(operation, "not_found", time.Since(start).Seconds())
		return []*model.Notification{}, nil
	}

	// Create pipeline for batch retrieval
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, id := range ids {
		key := fmt.Sprintf("%s%s", notificationPrefix, id)
		cmds[id] = pipe.Get(ctx, key)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return nil, fmt.Errorf("error retrieving notifications: %w", err)
	}

	// Process results
	notifications := make([]*model.Notification, 0, len(ids))
	for _, id := range ids {
		data, err := cmds[id].Bytes()
		if err != nil {
			if err != redis.Nil {
				r.logger.Error("error retrieving notification",
					zap.Error(err),
					zap.String("id", id),
				)
			}
			metrics.RecordCacheMiss()
			continue
		}

		metrics.RecordCacheHit()

		var notification model.Notification
		if err := json.Unmarshal(data, &notification); err != nil {
			r.logger.Error("error unmarshaling notification",
				zap.Error(err),
				zap.String("id", id),
			)
			continue
		}

		notifications = append(notifications, &notification)
	}

	metrics.RecordOperationDuration(operation, "success", time.Since(start).Seconds())
	return notifications, nil
}

// Update updates an existing notification
func (r *NotificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	start := time.Now()
	operation := "update"

	// Check if notification exists
	key := fmt.Sprintf("%s%s", notificationPrefix, notification.ID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error checking notification existence: %w", err)
	}
	if exists == 0 {
		metrics.RecordOperationDuration(operation, "not_found", time.Since(start).Seconds())
		return fmt.Errorf("notification not found: %s", notification.ID)
	}

	// Update notification
	data, err := json.Marshal(notification)
	if err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	if err := r.client.Set(ctx, key, data, defaultExpiration).Err(); err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error updating notification: %w", err)
	}

	metrics.RecordOperationDuration(operation, "success", time.Since(start).Seconds())
	metrics.UpdateNotificationStatus(string(notification.Status), 1)
	return nil
}

// DeleteByID deletes a notification by ID
func (r *NotificationRepository) DeleteByID(ctx context.Context, id string) error {
	start := time.Now()
	operation := "delete"

	notification, err := r.FindByID(ctx, id)
	if err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return err
	}
	if notification == nil {
		metrics.RecordOperationDuration(operation, "not_found", time.Since(start).Seconds())
		return nil // Already deleted
	}

	pipe := r.client.Pipeline()

	// Remove notification data
	notificationKey := fmt.Sprintf("%s%s", notificationPrefix, id)
	pipe.Del(ctx, notificationKey)

	// Remove from recipient's list
	recipientKey := fmt.Sprintf("%s%s", recipientPrefix, notification.Recipient)
	pipe.ZRem(ctx, recipientKey, id)

	if _, err := pipe.Exec(ctx); err != nil {
		metrics.RecordOperationDuration(operation, "error", time.Since(start).Seconds())
		return fmt.Errorf("error deleting notification: %w", err)
	}

	metrics.RecordOperationDuration(operation, "success", time.Since(start).Seconds())
	metrics.UpdateNotificationStatus(string(notification.Status), -1)
	return nil
}

// monitorRedisConnection periodically checks Redis connection status
func (r *NotificationRepository) monitorRedisConnection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := r.client.Ping(ctx).Err()
			metrics.SetRedisConnectionStatus(err == nil)
		}
	}
}
