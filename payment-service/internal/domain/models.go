package domain

import (
	"github.com/google/uuid"
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

type Account struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Payment struct {
	Id             uuid.UUID
	OrderId        uuid.UUID
	UserId         uuid.UUID
	AccountId      uuid.UUID
	Amount         float64
	PaymentMethod  string
	Status         PaymentStatus
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProcessedMessage struct {
	Id        uuid.UUID
	OrderId   uuid.UUID
	EventType string
	CreatedAt time.Time
}

