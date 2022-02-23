package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
)

// Withdrawal запрос на списание баллов с баланса для оплаты заказа OrderID
type Withdrawal struct {
	ID          string    `json:",omitempty"`
	UID         string    `json:",omitempty"`
	OrderID     string    `json:"order"`
	ProcessedAt time.Time `json:"processed_at"`
	// TODO https://github.com/shopspring/decimal
	Sum float32 `json:"sum"`
}

type WithdrawalOption func(session *Withdrawal)

func NewWithdrawal(userID string, orderID string, sum float32, opts ...WithdrawalOption) Withdrawal {
	withdrawal := Withdrawal{
		ID:          random.String(12),
		UID:         userID,
		OrderID:     orderID,
		ProcessedAt: time.Now(),
		Sum:         sum,
	}
	for _, opt := range opts {
		opt(&withdrawal)
	}
	return withdrawal
}
