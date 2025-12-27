package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/payment-service/internal/infrastructure/pgdb"
	"github.com/niklvrr/OnlineShop/payment-service/internal/usecase"
	paymentpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/payment"
	"log/slog"
)

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	paymentUC *usecase.PaymentUseCase
	logger    *slog.Logger
}

func NewPaymentHandler(paymentUC *usecase.PaymentUseCase, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentUC: paymentUC,
		logger:    logger,
	}
}

func (h *PaymentHandler) CreateAccount(ctx context.Context, req *paymentpb.CreateAccountRequest) (*paymentpb.CreateAccountResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	accountId, err := h.paymentUC.CreateAccount(ctx, userId)
	if err != nil {
		h.logger.Error("failed to create account", "error", err)
		return nil, err
	}

	return &paymentpb.CreateAccountResponse{
		AccountId: accountId.String(),
	}, nil
}

func (h *PaymentHandler) ReplenishAccount(ctx context.Context, req *paymentpb.ReplenishAccountRequest) (*paymentpb.ReplenishAccountResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, err
	}

	err = h.paymentUC.ReplenishAccount(ctx, accountId, float64(req.ReplenishAmount))
	if err != nil {
		h.logger.Error("failed to replenish account", "error", err)
		return nil, err
	}

	return &paymentpb.ReplenishAccountResponse{
		Status: "SUCCESS",
	}, nil
}

func (h *PaymentHandler) GetBalance(ctx context.Context, req *paymentpb.GetBalanceRequest) (*paymentpb.GetBalanceResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, err
	}

	balance, err := h.paymentUC.GetBalance(ctx, accountId)
	if err != nil {
		h.logger.Error("failed to get balance", "error", err)
		return nil, err
	}

	return &paymentpb.GetBalanceResponse{
		Balance: float32(balance),
	}, nil
}

func (h *PaymentHandler) DebitToAccount(ctx context.Context, req *paymentpb.DebitToAccountRequest) (*paymentpb.DebitToAccountResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, err
	}

	err = h.paymentUC.DebitAccount(ctx, accountId, float64(req.DebitAmount))
	if err != nil {
		h.logger.Error("failed to debit account", "error", err)
		return nil, err
	}

	return &paymentpb.DebitToAccountResponse{
		Status: "SUCCESS",
	}, nil
}

func (h *PaymentHandler) GetAccountByUserId(ctx context.Context, req *paymentpb.GetAccountByUserIdRequest) (*paymentpb.GetAccountByUserIdResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	accountId, err := h.paymentUC.GetAccountByUserId(ctx, userId)
	if err != nil {
		h.logger.Error("failed to get account by user id", "error", err)
		return nil, err
	}

	return &paymentpb.GetAccountByUserIdResponse{
		AccountId: accountId.String(),
	}, nil
}

func (h *PaymentHandler) InitiatePayment(ctx context.Context, req *paymentpb.InitiatePaymentRequest) (*paymentpb.InitiatePaymentResponse, error) {
	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, err
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}

	payment, err := h.paymentUC.InitiatePayment(ctx, orderId, userId, float64(req.Amount), req.PaymentMethod, idempotencyKey)
	if err != nil {
		if err == pgdb.ErrDuplicate {
			existingPayment, err := h.paymentUC.GetPaymentByOrderIdAndIdempotencyKey(ctx, orderId, idempotencyKey)
			if err != nil {
				h.logger.Error("failed to get existing payment", "error", err)
				return nil, err
			}
			return &paymentpb.InitiatePaymentResponse{
				PaymentId: existingPayment.Id.String(),
				Status:    string(existingPayment.Status),
			}, nil
		}
		h.logger.Error("failed to initiate payment", "error", err)
		return nil, err
	}

	return &paymentpb.InitiatePaymentResponse{
		PaymentId: payment.Id.String(),
		Status:    string(payment.Status),
	}, nil
}

func (h *PaymentHandler) GetPaymentStatus(ctx context.Context, req *paymentpb.GetPaymentStatusRequest) (*paymentpb.GetPaymentStatusResponse, error) {
	paymentId, err := uuid.Parse(req.PaymentId)
	if err != nil {
		return nil, err
	}

	payment, err := h.paymentUC.GetPaymentById(ctx, paymentId)
	if err != nil {
		h.logger.Error("failed to get payment status", "error", err)
		return nil, err
	}

	return &paymentpb.GetPaymentStatusResponse{
		PaymentId: payment.Id.String(),
		OrderId:   payment.OrderId.String(),
		Status:    string(payment.Status),
		Amount:    float32(payment.Amount),
	}, nil
}

