package entity

import (
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type Withdrawal struct {
	ID          string
	UID         string
	OrderID     string `json:"order"`
	ProcessedAt int64  `json:"processed_at"`
	// TODO https://github.com/shopspring/decimal
	Sum float32 `json:"sum"`
}

func NewWithdrawal(userID string, orderID string, sum float32) *Withdrawal {
	return &Withdrawal{
		ID:          random.String(12),
		UID:         userID,
		OrderID:     orderID,
		ProcessedAt: time.Now().UnixMilli(),
		Sum:         sum,
	}
}
