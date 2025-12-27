package usecase

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/order-service/internal/domain"
	"log/slog"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order, messagePayload []byte) error
	GetOrderById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error)
	GetAllOrders(ctx context.Context, userId uuid.UUID, offset, limit uint32) ([]*domain.Order, error)
	GetOrderStatus(ctx context.Context, orderId uuid.UUID) (*domain.Order, error)
}

type OrderUseCase struct {
	repo   OrderRepository
	logger *slog.Logger
}

func NewOrderUseCase(repo OrderRepository, logger *slog.Logger) *OrderUseCase {
	return &OrderUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, userId uuid.UUID, items []domain.OrderItem, metadata map[string]string) (*domain.Order, error) {
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price * float64(item.Qty)
	}

	order := &domain.Order{
		Id:       uuid.New(),
		UserId:   userId,
		Amount:   totalAmount,
		Status:   domain.OrderStatusNew,
		Items:    items,
		Metadata: metadata,
	}

	payload, err := json.Marshal(map[string]interface{}{
		"event_type": "OrderCreated",
		"order_id":   order.Id.String(),
		"user_id":    order.UserId.String(),
		"amount":     order.Amount,
		"items":      items,
	})
	if err != nil {
		return nil, err
	}

	err = uc.repo.CreateOrder(ctx, order, payload)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrderById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	return uc.repo.GetOrderById(ctx, orderId)
}

func (uc *OrderUseCase) GetAllOrders(ctx context.Context, userId uuid.UUID, page, limit uint32) ([]*domain.Order, error) {
	offset := (page - 1) * limit
	return uc.repo.GetAllOrders(ctx, userId, offset, limit)
}

func (uc *OrderUseCase) GetOrderStatus(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	return uc.repo.GetOrderStatus(ctx, orderId)
}

