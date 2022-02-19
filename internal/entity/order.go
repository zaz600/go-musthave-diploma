package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type OrderStatus string

const (
	OrderStatusNew            OrderStatus = "NEW"
	OrderStatusProcessing     OrderStatus = "PROCESSING"
	OrderStatusInvalid        OrderStatus = "INVALID"
	OrderStatusProcessed      OrderStatus = "PROCESSED"
	OrderStatusTooManyRetries OrderStatus = "TOO_MANY_RETRIES"
)

type Order struct {
	ID         string
	UID        string
	OrderID    string      `json:"number"`
	UploadedAt time.Time   `json:"uploaded_at"`
	Status     OrderStatus `json:"status"`
	// TODO https://github.com/shopspring/decimal
	Accrual float32 `json:"accrual"`

	RetryCount int
}

func NewOrder(userID string, orderID string) *Order {
	return &Order{
		ID:         random.String(12),
		UID:        userID,
		OrderID:    orderID,
		UploadedAt: time.Now(),
		Status:     OrderStatusNew,
		Accrual:    0,
		RetryCount: 0,
	}
}
