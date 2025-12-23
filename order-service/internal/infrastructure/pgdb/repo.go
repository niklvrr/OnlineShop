package pgdb

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/order-service/internal/domain"
)

const (
	createOrderQuery = `
INSERT INTO orders (user_id, amount, status)
VALUES ($1, $2, $3)
RETURNING id`

	addMessageToOutboxQuery = `
INSERT INTO outbox (order_id, event_type, payload, status)
VALUES ($1, $2, $3)
RETURNING id`

	getAllOrdersQuery = `
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2, OFFSET $3;`

	getOrderStatusQuery = `
SELECT status FROM orders
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
	// Начало транзакции
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return handleDBError(err)
	}

	commited := false
	defer func() {
		if !commited {
			tx.Rollback()
		}
	}()

	// Сохранение заказа в таблице заказов
	err = tx.QueryRowContext(ctx, createOrderQuery, order.UserId, order.Amount, order.Status).Scan(&order.Id)
	if err != nil {
		return handleDBError(err)
	}

	// Сохранение сообщения о заказе в таблце outbox
	_, err = tx.ExecContext(ctx, addMessageToOutboxQuery, order.Id, "OrderCreated", messagePayload, "PENDING")
	if err != nil {
		return handleDBError(err)
	}

	// Конец транзакции
	err = tx.Commit()
	if err != nil {
		return handleDBError(err)
	}
	commited = true
	return nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context, offset, limit uint32) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, getAllOrdersQuery, offset, limit)
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

func (r *OrderRepository) GetOrderStatus(ctx context.Context, orderId uuid.UUID) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, getOrderStatusQuery, orderId).Scan(&status)
	if err != nil {
		return "", handleDBError(err)
	}

	return status, nil
}
