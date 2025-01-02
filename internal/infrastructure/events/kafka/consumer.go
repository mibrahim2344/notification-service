package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/mibrahim2344/notification-service/internal/domain/services"
	"go.uber.org/zap"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	consumer        sarama.ConsumerGroup
	notificationSvc services.NotificationService
	logger          *zap.Logger
	topics          []string
	ready           chan bool
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(
	brokers []string,
	groupID string,
	topics []string,
	notificationSvc services.NotificationService,
	logger *zap.Logger,
) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumer:        consumer,
		notificationSvc: notificationSvc,
		logger:          logger,
		topics:          topics,
		ready:           make(chan bool),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start() error {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumer.Consume(c.ctx, c.topics, c); err != nil {
				c.logger.Error("error from consumer", zap.Error(err))
			}
			if c.ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready
	c.logger.Info("consumer is ready")

	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	c.cancel()
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("error closing consumer: %w", err)
	}
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			c.logger.Info("received message",
				zap.String("topic", message.Topic),
				zap.String("key", string(message.Key)),
				zap.Int64("offset", message.Offset),
				zap.Int32("partition", message.Partition),
			)

			if err := c.handleMessage(message); err != nil {
				c.logger.Error("error handling message",
					zap.Error(err),
					zap.String("topic", message.Topic),
					zap.Int64("offset", message.Offset),
				)
			}

			session.MarkMessage(message, "")

		case <-c.ctx.Done():
			return nil
		}
	}
}

func (c *Consumer) handleMessage(message *sarama.ConsumerMessage) error {
	// Extract event type from message key
	eventType := string(message.Key)

	// Handle the event using notification service
	if err := c.notificationSvc.HandleUserEvent(c.ctx, eventType, message.Value); err != nil {
		return fmt.Errorf("error handling user event: %w", err)
	}

	return nil
}
