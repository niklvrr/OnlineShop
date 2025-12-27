package handler

import (
	"context"
	"github.com/google/uuid"
	"github.com/niklvrr/OnlineShop/order-service/internal/domain"
	"github.com/niklvrr/OnlineShop/order-service/internal/usecase"
	orderpb "github.com/niklvrr/OnlineShop/proto-contracts/gen/go/order"
	"log/slog"
)

type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	orderUC *usecase.OrderUseCase
	logger  *slog.Logger
}

func NewOrderHandler(orderUC *usecase.OrderUseCase, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		orderUC: orderUC,
		logger:  logger,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItem{
			Sku:   item.Sku,
			Qty:   item.Qty,
			Price: float64(item.Price),
		})
	}

	metadata := make(map[string]string)
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	order, err := h.orderUC.CreateOrder(ctx, userId, items, metadata)
	if err != nil {
		h.logger.Error("failed to create order", "error", err)
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		OrderId: order.Id.String(),
	}, nil
}

func (h *OrderHandler) GetAllOrders(ctx context.Context, req *orderpb.GetAllOrdersRequest) (*orderpb.GetAllOrdersResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	orders, err := h.orderUC.GetAllOrders(ctx, uuid.Nil, page, limit)
	if err != nil {
		h.logger.Error("failed to get all orders", "error", err)
		return nil, err
	}

	pbOrders := make([]*orderpb.Order, 0, len(orders))
	for _, order := range orders {
		pbOrders = append(pbOrders, &orderpb.Order{
			Id:     order.Id.String(),
			UserId: order.UserId.String(),
			Amount: float32(order.Amount),
			Status: string(order.Status),
		})
	}

	return &orderpb.GetAllOrdersResponse{
		Orders: pbOrders,
	}, nil
}

func (h *OrderHandler) GetOrderStatus(ctx context.Context, req *orderpb.GetOrderStatusRequest) (*orderpb.GetOrderStatusResponse, error) {
	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, err
	}

	order, err := h.orderUC.GetOrderStatus(ctx, orderId)
	if err != nil {
		h.logger.Error("failed to get order status", "error", err)
		return nil, err
	}

	return &orderpb.GetOrderStatusResponse{
		OrderId: order.Id.String(),
		Status:  string(order.Status),
		Amount:  float32(order.Amount),
		UserId:  order.UserId.String(),
	}, nil
}

