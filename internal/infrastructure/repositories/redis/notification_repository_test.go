package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/mibrahim2344/notification-service/internal/domain/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type redisMock struct {
	*redis.Client
	commands map[string]func(args ...interface{}) (interface{}, error)
}

func setupTestRepo(t *testing.T) (*NotificationRepository, func()) {
	// Create a miniredis server
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client connected to miniredis
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	logger, _ := zap.NewDevelopment()
	repo := NewNotificationRepository(client, logger)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return repo, cleanup
}

func createTestNotification(recipient string) *model.Notification {
	return model.NewNotification(
		recipient,
		model.EmailNotification,
		model.EmailTemplate,
		uuid.New(),
		map[string]string{
			"testKey": "testValue",
		},
	)
}

func TestNotificationRepository_Save(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	notification := createTestNotification("test@example.com")

	// Test Save
	err := repo.Save(ctx, notification)
	require.NoError(t, err)

	// Verify saved notification
	saved, err := repo.FindByID(ctx, notification.ID.String())
	assert.NoError(t, err)
	if assert.NotNil(t, saved) {
		assert.Equal(t, notification.ID, saved.ID)
		assert.Equal(t, notification.Recipient, saved.Recipient)
		assert.Equal(t, notification.Subject, saved.Subject)
		assert.Equal(t, notification.Content, saved.Content)
		assert.Equal(t, notification.Type, saved.Type)
		assert.Equal(t, notification.Status, saved.Status)
	}
}

func TestNotificationRepository_FindByID(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	notification := createTestNotification("test@example.com")

	// Save test data
	err := repo.Save(ctx, notification)
	require.NoError(t, err)

	t.Run("Existing notification", func(t *testing.T) {
		found, err := repo.FindByID(ctx, notification.ID.String())
		assert.NoError(t, err)
		if assert.NotNil(t, found) {
			assert.Equal(t, notification.ID, found.ID)
			assert.Equal(t, notification.Content, found.Content)
		}
	})

	t.Run("Non-existing notification", func(t *testing.T) {
		found, err := repo.FindByID(ctx, uuid.New().String())
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestNotificationRepository_FindByRecipient(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	recipient := "test@example.com"

	// Create test notifications with explicit timestamps
	notifications := make([]*model.Notification, 5)
	baseTime := time.Now()
	
	for i := 0; i < 5; i++ {
		notification := createTestNotification(recipient)
		notification.CreatedAt = baseTime.Add(time.Duration(i) * time.Hour)
		err := repo.Save(ctx, notification)
		require.NoError(t, err)
		notifications[i] = notification
	}

	t.Run("Pagination test", func(t *testing.T) {
		// Test first page (should get the two most recent notifications)
		found, err := repo.FindByRecipient(ctx, recipient, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, found, 2)
		
		// Should be ordered by CreatedAt descending
		assert.True(t, found[0].CreatedAt.After(found[1].CreatedAt), "First notification should be more recent than second")
		assert.Equal(t, notifications[4].ID, found[0].ID, "Should get the most recent notification first")
		assert.Equal(t, notifications[3].ID, found[1].ID, "Should get the second most recent notification")

		// Test second page
		found, err = repo.FindByRecipient(ctx, recipient, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, found, 2)
		
		// Verify ordering continues
		assert.True(t, found[0].CreatedAt.After(found[1].CreatedAt), "First notification should be more recent than second")
		assert.Equal(t, notifications[2].ID, found[0].ID, "Should get the third most recent notification")
		assert.Equal(t, notifications[1].ID, found[1].ID, "Should get the fourth most recent notification")
	})

	t.Run("Empty result", func(t *testing.T) {
		found, err := repo.FindByRecipient(ctx, "nonexistent@example.com", 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, found)
	})
}

func TestNotificationRepository_Update(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	notification := createTestNotification("test@example.com")

	t.Run("Update existing notification", func(t *testing.T) {
		// Save initial notification
		err := repo.Save(ctx, notification)
		require.NoError(t, err)

		// Update notification
		notification.Status = model.StatusSent
		notification.Subject = "Updated Subject"
		err = repo.Update(ctx, notification)
		assert.NoError(t, err)

		// Verify update
		updated, err := repo.FindByID(ctx, notification.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, model.StatusSent, updated.Status)
		assert.Equal(t, "Updated Subject", updated.Subject)
	})

	t.Run("Update non-existing notification", func(t *testing.T) {
		nonExistent := createTestNotification("test@example.com")
		err := repo.Update(ctx, nonExistent)
		assert.Error(t, err)
	})
}

func TestNotificationRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	notification := createTestNotification("test@example.com")

	t.Run("Delete existing notification", func(t *testing.T) {
		// Save notification
		err := repo.Save(ctx, notification)
		require.NoError(t, err)

		// Delete notification
		err = repo.DeleteByID(ctx, notification.ID.String())
		assert.NoError(t, err)

		// Verify deletion
		found, err := repo.FindByID(ctx, notification.ID.String())
		assert.NoError(t, err)
		assert.Nil(t, found)

		// Verify removal from recipient's list
		notifications, err := repo.FindByRecipient(ctx, notification.Recipient, 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, notifications)
	})

	t.Run("Delete non-existing notification", func(t *testing.T) {
		err := repo.DeleteByID(ctx, uuid.New().String())
		assert.NoError(t, err) // Should not return error for non-existing notification
	})
}
