package entity

import (
	"github.com/ShiraazMoollatjie/goluhn"
)

// Account счет пользователя с аккумулированным балансом и суммой списаний
type Account struct {
	AccountID   string
	UID         string
	Balance     float32
	Withdrawals float32
}

type AccountOption func(*Account)

func NewAccount(userID string, opts ...AccountOption) Account {
	account := Account{
		AccountID:   goluhn.Generate(16),
		UID:         userID,
		Balance:     0,
		Withdrawals: 0,
	}
	for _, opt := range opts {
		opt(&account)
	}
	return account
}

func WithAccountID(accountID string) AccountOption {
	return func(s *Account) {
		s.AccountID = accountID
	}
}

func WithBalance(amount float32) AccountOption {
	return func(s *Account) {
		s.Balance = amount
	}
}

func WithWithdrawals(amount float32) AccountOption {
	return func(s *Account) {
		s.Withdrawals = amount
	}
}
