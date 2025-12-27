package pgdb

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/order-service/internal/domain"
	"time"
)

const (
	createOrderQuery = `
INSERT INTO orders (id, user_id, amount, status)
VALUES ($1, $2, $3, $4)
RETURNING created_at, updated_at`

	createOrderItemQuery = `
INSERT INTO order_items (id, order_id, sku, qty, price)
VALUES ($1, $2, $3, $4, $5)`

	addMessageToOutboxQuery = `
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, unique_key)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`

	getOrderByIdQuery = `
SELECT id, user_id, amount, status, created_at, updated_at
FROM orders
WHERE id = $1`

	getOrderItemsQuery = `
SELECT id, order_id, sku, qty, price, created_at
FROM order_items
WHERE order_id = $1`

	getAllOrdersQuery = `
SELECT id, user_id, amount, status, created_at, updated_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

	getOrderStatusQuery = `
SELECT id, user_id, amount, status
FROM orders
WHERE id = $1`

	getUnsentOutboxMessagesQuery = `
SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
FROM outbox
WHERE status = 'PENDING'
ORDER BY created_at ASC
LIMIT $1`

	markOutboxMessageAsSentQuery = `
UPDATE outbox
SET status = 'SENT', sent_at = now()
WHERE id = $1`
)

var ()

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order, messagePayload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return handleDBError(err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if order.Id == uuid.Nil {
		order.Id = uuid.New()
	}

	err = tx.QueryRowContext(ctx, createOrderQuery,
		order.Id, order.UserId, order.Amount, order.Status,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return handleDBError(err)
	}

	for i := range order.Items {
		if order.Items[i].Id == uuid.Nil {
			order.Items[i].Id = uuid.New()
		}
		order.Items[i].OrderId = order.Id
		_, err = tx.ExecContext(ctx, createOrderItemQuery,
			order.Items[i].Id, order.Items[i].OrderId,
			order.Items[i].Sku, order.Items[i].Qty, order.Items[i].Price,
		)
		if err != nil {
			return handleDBError(err)
		}
	}

	uniqueKey := order.Id.String() + "_OrderCreated"
	_, err = tx.ExecContext(ctx, addMessageToOutboxQuery,
		"order", order.Id, "OrderCreated", messagePayload, uniqueKey,
	)
	if err != nil {
		return handleDBError(err)
	}

	err = tx.Commit()
	if err != nil {
		return handleDBError(err)
	}
	committed = true
	return nil
}

func (r *OrderRepository) GetOrderById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, getOrderByIdQuery, orderId).Scan(
		&order.Id,
		&order.UserId,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}

	rows, err := r.db.QueryContext(ctx, getOrderItemsQuery, orderId)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	for rows.Next() {
		item := domain.OrderItem{}
		err := rows.Scan(
			&item.Id,
			&item.OrderId,
			&item.Sku,
			&item.Qty,
			&item.Price,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, userId uuid.UUID, offset, limit uint32) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, getAllOrdersQuery, userId, limit, offset)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.Id,
			&order.UserId,
			&order.Amount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, handleDBError(err)
	}

	return orders, nil
}

func (r *OrderRepository) GetOrderStatus(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, getOrderStatusQuery, orderId).Scan(
		&order.Id,
		&order.UserId,
		&order.Amount,
		&order.Status,
	)
	if err != nil {
		return nil, handleDBError(err)
	}

	return order, nil
}

func (r *OrderRepository) GetUnsentOutboxMessages(ctx context.Context, limit int) ([]OutboxMessage, error) {
	rows, err := r.db.QueryContext(ctx, getUnsentOutboxMessagesQuery, limit)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var messages []OutboxMessage
	for rows.Next() {
		msg := OutboxMessage{}
		err := rows.Scan(
			&msg.Id,
			&msg.AggregateType,
			&msg.AggregateId,
			&msg.EventType,
			&msg.Payload,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *OrderRepository) MarkOutboxMessageAsSent(ctx context.Context, messageId uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, markOutboxMessageAsSentQuery, messageId)
	if err != nil {
		return handleDBError(err)
	}
	return nil
}

type OutboxMessage struct {
	Id            uuid.UUID
	AggregateType string
	AggregateId   uuid.UUID
	EventType     string
	Payload       []byte
	CreatedAt     time.Time
}
