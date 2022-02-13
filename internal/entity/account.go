package entity

import "github.com/ShiraazMoollatjie/goluhn"

type Account struct {
	AccountID   string
	UID         string
	Balance     float32
	Withdrawals float32
}

func NewAccount(userID string) *Account {
	return &Account{
		AccountID:   goluhn.Generate(16),
		UID:         userID,
		Balance:     0,
		Withdrawals: 0,
	}
}
