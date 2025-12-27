package domain

import (
	"github.com/google/uuid"
	"time"
)

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "NEW"
	OrderStatusFinished  OrderStatus = "FINISHED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type OrderItem struct {
	Id        uuid.UUID
	OrderId   uuid.UUID
	Sku       string
	Qty       int32
	Price     float64
	CreatedAt time.Time
}

type Order struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Amount    float64
	Status    OrderStatus
	Items     []OrderItem
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}
