package model

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

type Order struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Amount    float64
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
