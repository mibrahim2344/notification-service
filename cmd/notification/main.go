package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mibrahim2344/notification-service/internal/application/notification"
	"github.com/mibrahim2344/notification-service/internal/infrastructure/events/kafka"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	config := loadConfig()

	// Initialize services
	notificationService := notification.NewService(
		// Initialize dependencies here
		nil, // repository
		nil, // email provider
		nil, // sms provider
		nil, // push provider
		nil, // template engine
		logger,
	)

	// Create Kafka consumer
	consumer, err := kafka.NewConsumer(
		config.Kafka.Brokers,
		config.Kafka.GroupID,
		[]string{config.Kafka.Topics.UserEvents},
		notificationService,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create consumer", zap.Error(err))
	}

	// Start consumer
	if err := consumer.Start(); err != nil {
		logger.Fatal("failed to start consumer", zap.Error(err))
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("received shutdown signal")

	if err := consumer.Stop(); err != nil {
		logger.Error("error stopping consumer", zap.Error(err))
	}
}

type Config struct {
	Kafka struct {
		Brokers []string
		GroupID string
		Topics  struct {
			UserEvents string
		}
	}
}

func loadConfig() *Config {
	// TODO: Implement configuration loading
	return &Config{
		Kafka: struct {
			Brokers []string
			GroupID string
			Topics  struct {
				UserEvents string
			}
		}{
			Brokers: []string{"localhost:9092"},
			GroupID: "notification-service",
			Topics: struct {
				UserEvents string
			}{
				UserEvents: "user-events",
			},
		},
	}
}
