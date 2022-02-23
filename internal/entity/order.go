package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
)

type OrderStatus string

const (
	OrderStatusNew            OrderStatus = "NEW"
	OrderStatusProcessing     OrderStatus = "PROCESSING"
	OrderStatusInvalid        OrderStatus = "INVALID"
	OrderStatusProcessed      OrderStatus = "PROCESSED"
	OrderStatusTooManyRetries OrderStatus = "TOO_MANY_RETRIES"
)

// Order заказ, загружаемый пользователем, за который могут быть начислены баллы лояльности
type Order struct {
	ID         string      `json:",omitempty"`
	UID        string      `json:",omitempty"`
	OrderID    string      `json:"number"`
	UploadedAt time.Time   `json:"uploaded_at"`
	Status     OrderStatus `json:"status"`
	// TODO https://github.com/shopspring/decimal
	Accrual float32 `json:"accrual"`

	// RetryCount количество попыток получить начисления во внешнем сервисе
	RetryCount int `json:",omitempty"`
}

type OrderOption func(*Order)

func NewOrder(userID string, orderID string, opts ...OrderOption) Order {
	order := Order{
		ID:         random.String(12),
		UID:        userID,
		OrderID:    orderID,
		UploadedAt: time.Now(),
		Status:     OrderStatusNew,
		Accrual:    0,
		RetryCount: 0,
	}
	for _, opt := range opts {
		opt(&order)
	}
	return order
}

func WithStatus(status OrderStatus) OrderOption {
	return func(o *Order) {
		o.Status = status
	}
}

func WithAccrual(accrual float32) OrderOption {
	return func(o *Order) {
		o.Accrual = accrual
	}
}

func WithUploadedAt(uploadedAt time.Time) OrderOption {
	return func(o *Order) {
		o.UploadedAt = uploadedAt
	}
}
