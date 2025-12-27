package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/payment-service/internal/usecase"
	"log/slog"
	"sync"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	paymentUC     *usecase.PaymentUseCase
	logger        *slog.Logger
}

func NewConsumer(brokers []string, groupID string, paymentUC *usecase.PaymentUseCase, logger *slog.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		paymentUC:     paymentUC,
		logger:        logger,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, topics []string) error {
	handler := &consumerGroupHandler{
		paymentUC: c.paymentUC,
		logger:    c.logger,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
				c.logger.Error("error from consumer", "error", err)
				return
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-ctx.Done()
	c.logger.Info("shutting down consumer")
	wg.Wait()

	return c.consumerGroup.Close()
}

type consumerGroupHandler struct {
	paymentUC *usecase.PaymentUseCase
	logger    *slog.Logger
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var event map[string]interface{}
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.logger.Error("failed to unmarshal event", "error", err)
				session.MarkMessage(message, "")
				continue
			}

			eventType, ok := event["event_type"].(string)
			if !ok {
				h.logger.Error("event_type not found in event")
				session.MarkMessage(message, "")
				continue
			}

			switch eventType {
			case "OrderCreated":
				orderIdStr, _ := event["order_id"].(string)
				userIdStr, _ := event["user_id"].(string)
				amount, _ := event["amount"].(float64)

				orderId, err := uuid.Parse(orderIdStr)
				if err != nil {
					h.logger.Error("failed to parse order_id", "error", err)
					session.MarkMessage(message, "")
					continue
				}

				userId, err := uuid.Parse(userIdStr)
				if err != nil {
					h.logger.Error("failed to parse user_id", "error", err)
					session.MarkMessage(message, "")
					continue
				}

				err = h.paymentUC.ProcessOrderCreatedEvent(context.Background(), orderId, userId, amount)
				if err != nil {
					h.logger.Error("failed to process OrderCreated event", "error", err)
					session.MarkMessage(message, "")
					continue
				}

				h.logger.Info("processed OrderCreated event", "order_id", orderId)
			default:
				h.logger.Debug("unknown event type", "event_type", eventType)
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

