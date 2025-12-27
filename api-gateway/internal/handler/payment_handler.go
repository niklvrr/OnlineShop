package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/niklvrr/OnlineShop/api-gateway/internal/client"
	paymentpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/payment"
	"log/slog"
	"net/http"
)

type PaymentHandler struct {
	paymentClient *client.PaymentClient
	logger        *slog.Logger
}

func NewPaymentHandler(paymentClient *client.PaymentClient, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentClient: paymentClient,
		logger:        logger,
	}
}

func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderId        string `json:"order_id"`
		UserId         string `json:"user_id"`
		Amount         float64 `json:"amount"`
		PaymentMethod  string `json:"payment_method"`
		IdempotencyKey string `json:"idempotency_key,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	grpcReq := &paymentpb.InitiatePaymentRequest{
		OrderId:        req.OrderId,
		UserId:         req.UserId,
		Amount:         float32(req.Amount),
		PaymentMethod:  req.PaymentMethod,
		IdempotencyKey: req.IdempotencyKey,
	}

	resp, err := h.paymentClient.InitiatePayment(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to initiate payment", "error", err)
		http.Error(w, "failed to initiate payment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id": resp.PaymentId,
		"status":     resp.Status,
	})
}

func (h *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentId := vars["payment_id"]

	grpcReq := &paymentpb.GetPaymentStatusRequest{
		PaymentId: paymentId,
	}

	resp, err := h.paymentClient.GetPaymentStatus(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to get payment status", "error", err)
		http.Error(w, "failed to get payment status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id": resp.PaymentId,
		"order_id":   resp.OrderId,
		"status":     resp.Status,
		"amount":     resp.Amount,
	})
}

