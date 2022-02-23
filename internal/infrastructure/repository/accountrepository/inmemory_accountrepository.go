package accountrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

var ErrAccountExists = errors.New("account already exists")
var ErrUserAccountNotFound = errors.New("user account not found")

type InmemoryAccountRepository struct {
	mu           *sync.RWMutex
	db           map[string]entity.Account
	userAccounts map[string]entity.Account
}

func (r InmemoryAccountRepository) GetAccount(_ context.Context, userID string) (entity.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if account, ok := r.userAccounts[userID]; ok {
		return account, nil
	}
	return entity.Account{}, ErrUserAccountNotFound
}

func (r InmemoryAccountRepository) AddAccount(_ context.Context, account entity.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[account.AccountID]; ok {
		return ErrAccountExists
	}
	r.db[account.AccountID] = account
	r.userAccounts[account.UID] = account
	return nil
}

func (r InmemoryAccountRepository) RefillAmount(_ context.Context, userID string, diff float32) error {
	if diff <= 0 {
		return fmt.Errorf("invalid refill amount. Must be >= 0")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if account, ok := r.userAccounts[userID]; ok {
		account.Balance += diff
		r.db[account.AccountID] = account
		r.userAccounts[account.UID] = account
		return nil
	}
	return ErrUserAccountNotFound
}

func (r InmemoryAccountRepository) WithdrawalAmount(_ context.Context, userID string, diff float32) error {
	if diff <= 0 {
		return fmt.Errorf("invalid withdrawal amount. Must be >= 0")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if account, ok := r.userAccounts[userID]; ok {
		account.Balance -= diff
		account.Withdrawals += diff
		r.db[account.AccountID] = account
		r.userAccounts[account.UID] = account
		return nil
	}
	return ErrUserAccountNotFound
}

func (r InmemoryAccountRepository) Close() error {
	return nil
}

func NewInmemoryAccountRepository() *InmemoryAccountRepository {
	return &InmemoryAccountRepository{
		mu:           &sync.RWMutex{},
		db:           make(map[string]entity.Account, 10),
		userAccounts: make(map[string]entity.Account, 10),
	}
}
