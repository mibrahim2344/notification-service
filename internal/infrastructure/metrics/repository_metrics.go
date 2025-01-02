package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RepositoryOperationDuration tracks the duration of repository operations
	RepositoryOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "notification_repository_operation_duration_seconds",
			Help: "Duration of repository operations in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation", "status"},
	)

	// RepositoryOperationTotal tracks the total number of repository operations
	RepositoryOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_repository_operation_total",
			Help: "Total number of repository operations",
		},
		[]string{"operation", "status"},
	)

	// NotificationStorageSize tracks the size of stored notifications
	NotificationStorageSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notification_storage_size_bytes",
			Help: "Size of stored notifications in bytes",
		},
		[]string{"type"},
	)

	// NotificationsByStatus tracks the number of notifications by status
	NotificationsByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notifications_by_status_total",
			Help: "Number of notifications by status",
		},
		[]string{"status"},
	)

	// RedisConnectionStatus tracks the Redis connection status
	RedisConnectionStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_redis_connection_status",
			Help: "Status of Redis connection (1 for connected, 0 for disconnected)",
		},
	)

	// RedisCacheHits tracks Redis cache hits and misses
	RedisCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_redis_cache_hits_total",
			Help: "Number of Redis cache hits and misses",
		},
		[]string{"type"}, // hit or miss
	)
)

// RecordOperationDuration records the duration of a repository operation
func RecordOperationDuration(operation string, status string, duration float64) {
	RepositoryOperationDuration.WithLabelValues(operation, status).Observe(duration)
	RepositoryOperationTotal.WithLabelValues(operation, status).Inc()
}

// UpdateNotificationStorageSize updates the size of stored notifications
func UpdateNotificationStorageSize(notificationType string, sizeBytes float64) {
	NotificationStorageSize.WithLabelValues(notificationType).Set(sizeBytes)
}

// UpdateNotificationStatus updates the count of notifications by status
func UpdateNotificationStatus(status string, count float64) {
	NotificationsByStatus.WithLabelValues(status).Set(count)
}

// SetRedisConnectionStatus sets the Redis connection status
func SetRedisConnectionStatus(connected bool) {
	if connected {
		RedisConnectionStatus.Set(1)
	} else {
		RedisConnectionStatus.Set(0)
	}
}

// RecordCacheHit records a Redis cache hit
func RecordCacheHit() {
	RedisCacheHits.WithLabelValues("hit").Inc()
}

// RecordCacheMiss records a Redis cache miss
func RecordCacheMiss() {
	RedisCacheHits.WithLabelValues("miss").Inc()
}
