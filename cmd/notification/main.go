package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strconv"

	"github.com/mibrahim2344/notification-service/internal/api/handlers"
	apiservices "github.com/mibrahim2344/notification-service/internal/api/services"
	"github.com/mibrahim2344/notification-service/internal/application/notification"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/db"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/repositories/postgres"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize database connection
	dbConfig := db.DefaultConfig()
	dbConfig.Host = getEnv("DB_HOST", dbConfig.Host)
	dbConfig.Port = getEnvAsInt("DB_PORT", dbConfig.Port)
	dbConfig.User = getEnv("DB_USER", dbConfig.User)
	dbConfig.Password = getEnv("DB_PASSWORD", dbConfig.Password)
	dbConfig.DBName = getEnv("DB_NAME", dbConfig.DBName)
	dbConfig.SSLMode = getEnv("DB_SSLMODE", dbConfig.SSLMode)

	// Configure connection pool based on environment
	dbConfig.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", dbConfig.MaxOpenConns)
	dbConfig.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", dbConfig.MaxIdleConns)
	dbConfig.ConnMaxLifetime = getEnvAsDuration("DB_CONN_MAX_LIFETIME", dbConfig.ConnMaxLifetime)
	dbConfig.ConnMaxIdleTime = getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", dbConfig.ConnMaxIdleTime)

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close(database)

	// Initialize health checker
	healthChecker := db.NewHealthChecker(database, 30*time.Second, 5*time.Second)
	healthChecker.Start()
	defer healthChecker.Stop()

	// Initialize repositories
	notificationRepo := postgres.NewNotificationRepository(database)
	templateRepo := postgres.NewTemplateRepository(database)

	// Initialize services
	notificationService := notification.NewService(
		notificationRepo,
		nil, // email provider
		nil, // sms provider
		nil, // push provider
		templateRepo,
		logger,
	)

	// Initialize adapter and handlers
	notificationServiceAdapter := apiservices.NewNotificationServiceAdapter(notificationService)
	notificationHandler := handlers.NewNotificationHandler(notificationServiceAdapter, logger)

	// Initialize HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      setupRoutes(notificationHandler),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func setupRoutes(notificationHandler *handlers.NotificationHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/notifications", notificationHandler.SendNotification)
	return mux
}
