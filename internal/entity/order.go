package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type OrderStatus string

const (
	NEW        OrderStatus = "NEW"
	PROCESSING OrderStatus = "PROCESSING"
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         string
	UID        string
	OrderID    string      `json:"number"`
	UploadedAt int64       `json:"uploaded_at"`
	Status     OrderStatus `json:"status"`
	Accrual    int         `json:"accrual"`
}

func NewOrder(userID string, orderID string) *Order {
	return &Order{
		ID:         random.String(12),
		UID:        userID,
		OrderID:    orderID,
		UploadedAt: time.Now().UnixMilli(),
		Status:     NEW,
		Accrual:    0,
	}
}
