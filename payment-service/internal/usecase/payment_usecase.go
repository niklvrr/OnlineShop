package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/payment-service/internal/domain"
	"github.com/niklvrr/OnlineShop/payment-service/internal/infrastructure/pgdb"
	"log/slog"
)

type PaymentRepository interface {
	CreateAccount(ctx context.Context, account *domain.Account) error
	GetAccountByUserId(ctx context.Context, userId uuid.UUID) (*domain.Account, error)
	GetAccountById(ctx context.Context, accountId uuid.UUID) (*domain.Account, error)
	UpdateAccountBalance(ctx context.Context, accountId uuid.UUID, amount float64) (float64, error)
	CreatePayment(ctx context.Context, payment *domain.Payment, messagePayload []byte) error
	GetPaymentById(ctx context.Context, paymentId uuid.UUID) (*domain.Payment, error)
	GetPaymentByOrderIdAndIdempotencyKey(ctx context.Context, orderId uuid.UUID, idempotencyKey string) (*domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, paymentId uuid.UUID, status domain.PaymentStatus) error
	IsMessageProcessed(ctx context.Context, orderId uuid.UUID, eventType string) (bool, error)
	MarkMessageAsProcessed(ctx context.Context, orderId uuid.UUID, eventType string) error
}

type PaymentUseCase struct {
	repo   PaymentRepository
	logger *slog.Logger
}

func NewPaymentUseCase(repo PaymentRepository, logger *slog.Logger) *PaymentUseCase {
	return &PaymentUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (uc *PaymentUseCase) CreateAccount(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {
	existingAccount, err := uc.repo.GetAccountByUserId(ctx, userId)
	if err == nil && existingAccount != nil {
		return existingAccount.Id, nil
	}
	if err != nil && err != pgdb.ErrNotFound {
		return uuid.Nil, err
	}

	account := &domain.Account{
		Id:      uuid.New(),
		UserId:  userId,
		Balance: 0.0,
	}

	err = uc.repo.CreateAccount(ctx, account)
	if err != nil {
		return uuid.Nil, err
	}

	return account.Id, nil
}

func (uc *PaymentUseCase) ReplenishAccount(ctx context.Context, accountId uuid.UUID, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	_, err := uc.repo.UpdateAccountBalance(ctx, accountId, amount)
	if err != nil {
		return err
	}

	return nil
}

func (uc *PaymentUseCase) GetBalance(ctx context.Context, accountId uuid.UUID) (float64, error) {
	account, err := uc.repo.GetAccountById(ctx, accountId)
	if err != nil {
		return 0, err
	}

	return account.Balance, nil
}

func (uc *PaymentUseCase) DebitAccount(ctx context.Context, accountId uuid.UUID, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	account, err := uc.repo.GetAccountById(ctx, accountId)
	if err != nil {
		return err
	}

	if account.Balance < amount {
		return errors.New("insufficient balance")
	}

	_, err = uc.repo.UpdateAccountBalance(ctx, accountId, -amount)
	if err != nil {
		return err
	}

	return nil
}

func (uc *PaymentUseCase) GetAccountByUserId(ctx context.Context, userId uuid.UUID) (uuid.UUID, error) {
	account, err := uc.repo.GetAccountByUserId(ctx, userId)
	if err != nil {
		return uuid.Nil, err
	}

	return account.Id, nil
}

func (uc *PaymentUseCase) InitiatePayment(ctx context.Context, orderId, userId uuid.UUID, amount float64, paymentMethod, idempotencyKey string) (*domain.Payment, error) {
	existingPayment, err := uc.repo.GetPaymentByOrderIdAndIdempotencyKey(ctx, orderId, idempotencyKey)
	if err == nil && existingPayment != nil {
		return existingPayment, pgdb.ErrDuplicate
	}
	if err != nil && err != pgdb.ErrNotFound {
		return nil, err
	}

	account, err := uc.repo.GetAccountByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	if account.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	payment := &domain.Payment{
		Id:             uuid.New(),
		OrderId:        orderId,
		UserId:         userId,
		AccountId:      account.Id,
		Amount:         amount,
		PaymentMethod:  paymentMethod,
		Status:         domain.PaymentStatusPending,
		IdempotencyKey: idempotencyKey,
	}

	payload, err := json.Marshal(map[string]interface{}{
		"payment_id": payment.Id.String(),
		"order_id":   payment.OrderId.String(),
		"user_id":    payment.UserId.String(),
		"amount":     payment.Amount,
		"status":     payment.Status,
	})
	if err != nil {
		return nil, err
	}

	err = uc.repo.CreatePayment(ctx, payment, payload)
	if err != nil {
		return nil, err
	}

	_, err = uc.repo.UpdateAccountBalance(ctx, account.Id, -amount)
	if err != nil {
		return nil, err
	}

	payment.Status = domain.PaymentStatusCompleted
	err = uc.repo.UpdatePaymentStatus(ctx, payment.Id, domain.PaymentStatusCompleted)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentById(ctx context.Context, paymentId uuid.UUID) (*domain.Payment, error) {
	return uc.repo.GetPaymentById(ctx, paymentId)
}

func (uc *PaymentUseCase) GetPaymentByOrderIdAndIdempotencyKey(ctx context.Context, orderId uuid.UUID, idempotencyKey string) (*domain.Payment, error) {
	return uc.repo.GetPaymentByOrderIdAndIdempotencyKey(ctx, orderId, idempotencyKey)
}

func (uc *PaymentUseCase) ProcessOrderCreatedEvent(ctx context.Context, orderId uuid.UUID, userId uuid.UUID, amount float64) error {
	processed, err := uc.repo.IsMessageProcessed(ctx, orderId, "OrderCreated")
	if err != nil {
		return err
	}
	if processed {
		uc.logger.Debug("message already processed", "order_id", orderId, "event_type", "OrderCreated")
		return nil
	}

	_, err = uc.CreateAccount(ctx, userId)
	if err != nil {
		return err
	}

	err = uc.repo.MarkMessageAsProcessed(ctx, orderId, "OrderCreated")
	if err != nil {
		return err
	}

	return nil
}

