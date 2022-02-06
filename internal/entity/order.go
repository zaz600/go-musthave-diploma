package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type OrderStatus string

const (
	OrderStatusNEW            OrderStatus = "NEW"
	OrderStatusPROCESSING     OrderStatus = "PROCESSING"
	OrderStatusINVALID        OrderStatus = "INVALID"
	OrderStatusPROCESSED      OrderStatus = "PROCESSED"
	OrderStatusTECHNICALERROR OrderStatus = "TECHNICAL_ERROR"
	OrderStatusTooManyRetries OrderStatus = "TOO_MANY_RETRIES"
)

type TaskContext struct {
	RetryCount  int
	NextRetryAt time.Time
}

type Order struct {
	ID         string
	UID        string
	OrderID    string      `json:"number"`
	UploadedAt int64       `json:"uploaded_at"`
	Status     OrderStatus `json:"status"`
	// TODO https://github.com/shopspring/decimal
	Accrual float32 `json:"accrual"`

	Context TaskContext
}

func NewOrder(userID string, orderID string) *Order {
	return &Order{
		ID:         random.String(12),
		UID:        userID,
		OrderID:    orderID,
		UploadedAt: time.Now().UnixMilli(),
		Status:     OrderStatusNEW,
		Accrual:    0,
		Context: TaskContext{
			RetryCount:  0,
			NextRetryAt: time.Now(),
		},
	}
}
