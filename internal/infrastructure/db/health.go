package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dbHealthGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "notification_db_health_status",
		Help: "Database health status (1 for healthy, 0 for unhealthy)",
	})
	dbConnectionGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "notification_db_connections",
		Help: "Database connection pool statistics",
	}, []string{"state"})
)

// HealthChecker monitors database health
type HealthChecker struct {
	db        *sql.DB
	interval  time.Duration
	timeout   time.Duration
	stopChan  chan struct{}
	stopOnce  sync.Once
	isHealthy bool
	mu        sync.RWMutex
}

// NewHealthChecker creates a new database health checker
func NewHealthChecker(db *sql.DB, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		db:       db,
		interval: interval,
		timeout:  timeout,
		stopChan: make(chan struct{}),
	}
}

// Start starts the health checker
func (h *HealthChecker) Start() {
	go h.monitor()
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	h.stopOnce.Do(func() {
		close(h.stopChan)
	})
}

// IsHealthy returns the current health status
func (h *HealthChecker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isHealthy
}

// monitor periodically checks database health
func (h *HealthChecker) monitor() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.checkHealth()
			h.updateMetrics()
		}
	}
}

// checkHealth performs a health check
func (h *HealthChecker) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	err := h.db.PingContext(ctx)

	h.mu.Lock()
	h.isHealthy = err == nil
	h.mu.Unlock()

	if err != nil {
		fmt.Printf("Database health check failed: %v\n", err)
	}

	if h.isHealthy {
		dbHealthGauge.Set(1)
	} else {
		dbHealthGauge.Set(0)
	}
}

// updateMetrics updates Prometheus metrics
func (h *HealthChecker) updateMetrics() {
	stats := h.db.Stats()

	dbConnectionGauge.WithLabelValues("open").Set(float64(stats.OpenConnections))
	dbConnectionGauge.WithLabelValues("in_use").Set(float64(stats.InUse))
	dbConnectionGauge.WithLabelValues("idle").Set(float64(stats.Idle))
	dbConnectionGauge.WithLabelValues("max_open").Set(float64(stats.MaxOpenConnections))
}
