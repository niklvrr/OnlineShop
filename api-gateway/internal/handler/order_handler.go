package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/niklvrr/OnlineShop/api-gateway/internal/client"
	orderpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/order"
	"log/slog"
	"net/http"
)

type OrderHandler struct {
	orderClient *client.OrderClient
	logger      *slog.Logger
}

func NewOrderHandler(orderClient *client.OrderClient, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		orderClient: orderClient,
		logger:      logger,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserId   string                 `json:"user_id"`
		Items    []OrderItemRequest      `json:"items"`
		Metadata map[string]string      `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	items := make([]*orderpb.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &orderpb.OrderItem{
			Sku:   item.Sku,
			Qty:   item.Qty,
			Price: float32(item.Price),
		})
	}

	grpcReq := &orderpb.CreateOrderRequest{
		UserId:   req.UserId,
		Items:    items,
		Metadata: req.Metadata,
	}

	resp, err := h.orderClient.CreateOrder(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to create order", "error", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": resp.GetOrderId(),
	})
}

func (h *OrderHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := vars["order_id"]

	grpcReq := &orderpb.GetOrderStatusRequest{
		OrderId: orderId,
	}

	resp, err := h.orderClient.GetOrderStatus(r.Context(), grpcReq)
	if err != nil {
		h.logger.Error("failed to get order status", "error", err)
		http.Error(w, "failed to get order status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": resp.GetOrderId(),
		"status":   resp.GetStatus(),
		"amount":   resp.GetAmount(),
		"user_id":  resp.GetUserId(),
	})
}

type OrderItemRequest struct {
	Sku   string  `json:"sku"`
	Qty   int32   `json:"qty"`
	Price float64 `json:"price"`
}

