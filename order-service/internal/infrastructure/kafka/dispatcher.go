package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/niklvrr/OnlineShop/order-service/internal/infrastructure/pgdb"
	"log/slog"
	"time"
)

type Dispatcher struct {
	producer  sarama.SyncProducer
	repo      *pgdb.OrderRepository
	logger    *slog.Logger
	topic     string
	interval  time.Duration
	batchSize int
}

func NewDispatcher(brokers []string, repo *pgdb.OrderRepository, logger *slog.Logger, topic string) (*Dispatcher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Dispatcher{
		producer:  producer,
		repo:      repo,
		logger:    logger,
		topic:     topic,
		interval:  5 * time.Second,
		batchSize: 100,
	}, nil
}

func (d *Dispatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("stopping outbox dispatcher")
			return
		case <-ticker.C:
			d.processOutbox(ctx)
		}
	}
}

func (d *Dispatcher) processOutbox(ctx context.Context) {
	messages, err := d.repo.GetUnsentOutboxMessages(ctx, d.batchSize)
	if err != nil {
		d.logger.Error("failed to get unsent messages", "error", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	d.logger.Debug("processing outbox messages", "count", len(messages))

	for _, msg := range messages {
		if err := d.publishMessage(ctx, msg); err != nil {
			d.logger.Error("failed to publish message", "message_id", msg.Id, "error", err)
			continue
		}

		if err := d.repo.MarkOutboxMessageAsSent(ctx, msg.Id); err != nil {
			d.logger.Error("failed to mark message as sent", "message_id", msg.Id, "error", err)
			continue
		}

		d.logger.Debug("message published and marked as sent", "message_id", msg.Id, "event_type", msg.EventType)
	}
}

func (d *Dispatcher) publishMessage(ctx context.Context, msg pgdb.OutboxMessage) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic: d.topic,
		Value: sarama.ByteEncoder(msg.Payload),
		Key:   sarama.StringEncoder(msg.AggregateId.String()),
	}

	partition, offset, err := d.producer.SendMessage(kafkaMsg)
	if err != nil {
		return err
	}

	d.logger.Debug("message published",
		"message_id", msg.Id,
		"partition", partition,
		"offset", offset,
		"event_type", msg.EventType,
	)

	return nil
}

func (d *Dispatcher) Close() error {
	return d.producer.Close()
}

