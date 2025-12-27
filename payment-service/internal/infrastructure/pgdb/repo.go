package pgdb

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/payment-service/internal/domain"
)

const (
	createAccountQuery = `
INSERT INTO accounts (id, user_id, balance)
VALUES ($1, $2, $3)
RETURNING created_at, updated_at`

	getAccountByUserIdQuery = `
SELECT id, user_id, balance, created_at, updated_at
FROM accounts
WHERE user_id = $1`

	getAccountByIdQuery = `
SELECT id, user_id, balance, created_at, updated_at
FROM accounts
WHERE id = $1`

	updateAccountBalanceQuery = `
UPDATE accounts
SET balance = balance + $1, updated_at = now()
WHERE id = $2
RETURNING balance`

	createPaymentQuery = `
INSERT INTO payments (id, order_id, user_id, account_id, amount, payment_method, status, idempotency_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING created_at, updated_at`

	getPaymentByIdQuery = `
SELECT id, order_id, user_id, account_id, amount, payment_method, status, idempotency_key, created_at, updated_at
FROM payments
WHERE id = $1`

	getPaymentByOrderIdAndIdempotencyKeyQuery = `
SELECT id, order_id, user_id, account_id, amount, payment_method, status, idempotency_key, created_at, updated_at
FROM payments
WHERE order_id = $1 AND idempotency_key = $2`

	updatePaymentStatusQuery = `
UPDATE payments
SET status = $1, updated_at = now()
WHERE id = $2`

	addMessageToOutboxQuery = `
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, unique_key)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`

	checkProcessedMessageQuery = `
SELECT id FROM processed_messages
WHERE order_id = $1 AND event_type = $2`

	markMessageAsProcessedQuery = `
INSERT INTO processed_messages (order_id, event_type)
VALUES ($1, $2)
ON CONFLICT (order_id, event_type) DO NOTHING
RETURNING id`
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

func (r *PaymentRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	err := r.db.QueryRowContext(ctx, createAccountQuery,
		account.Id, account.UserId, account.Balance,
	).Scan(&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return handleDBError(err)
	}
	return nil
}

func (r *PaymentRepository) GetAccountByUserId(ctx context.Context, userId uuid.UUID) (*domain.Account, error) {
	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, getAccountByUserIdQuery, userId).Scan(
		&account.Id,
		&account.UserId,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}
	return account, nil
}

func (r *PaymentRepository) GetAccountById(ctx context.Context, accountId uuid.UUID) (*domain.Account, error) {
	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, getAccountByIdQuery, accountId).Scan(
		&account.Id,
		&account.UserId,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}
	return account, nil
}

func (r *PaymentRepository) UpdateAccountBalance(ctx context.Context, accountId uuid.UUID, amount float64) (float64, error) {
	var balance float64
	err := r.db.QueryRowContext(ctx, updateAccountBalanceQuery, amount, accountId).Scan(&balance)
	if err != nil {
		return 0, handleDBError(err)
	}
	return balance, nil
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment, messagePayload []byte) error {
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

	err = tx.QueryRowContext(ctx, createPaymentQuery,
		payment.Id, payment.OrderId, payment.UserId, payment.AccountId,
		payment.Amount, payment.PaymentMethod, payment.Status, payment.IdempotencyKey,
	).Scan(&payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return handleDBError(err)
	}

	uniqueKey := payment.OrderId.String() + "_" + payment.IdempotencyKey
	_, err = tx.ExecContext(ctx, addMessageToOutboxQuery,
		"payment", payment.Id, "PaymentInitiated", messagePayload, uniqueKey,
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

func (r *PaymentRepository) GetPaymentById(ctx context.Context, paymentId uuid.UUID) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, getPaymentByIdQuery, paymentId).Scan(
		&payment.Id,
		&payment.OrderId,
		&payment.UserId,
		&payment.AccountId,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}
	return payment, nil
}

func (r *PaymentRepository) GetPaymentByOrderIdAndIdempotencyKey(ctx context.Context, orderId uuid.UUID, idempotencyKey string) (*domain.Payment, error) {
	payment := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, getPaymentByOrderIdAndIdempotencyKeyQuery, orderId, idempotencyKey).Scan(
		&payment.Id,
		&payment.OrderId,
		&payment.UserId,
		&payment.AccountId,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}
	return payment, nil
}

func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, paymentId uuid.UUID, status domain.PaymentStatus) error {
	_, err := r.db.ExecContext(ctx, updatePaymentStatusQuery, status, paymentId)
	if err != nil {
		return handleDBError(err)
	}
	return nil
}

func (r *PaymentRepository) IsMessageProcessed(ctx context.Context, orderId uuid.UUID, eventType string) (bool, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, checkProcessedMessageQuery, orderId, eventType).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, handleDBError(err)
	}
	return true, nil
}

func (r *PaymentRepository) MarkMessageAsProcessed(ctx context.Context, orderId uuid.UUID, eventType string) error {
	_, err := r.db.ExecContext(ctx, markMessageAsProcessedQuery, orderId, eventType)
	if err != nil {
		return handleDBError(err)
	}
	return nil
}

