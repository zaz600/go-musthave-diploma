package withdrawalrepository

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemorWithdrawalRepository struct {
	mu              sync.RWMutex
	db              map[string]*entity.Withdrawal
	userWithdrawals map[string][]*entity.Withdrawal
}

func (r *InmemorWithdrawalRepository) AddWithdrawal(ctx context.Context, withdrawal *entity.Withdrawal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if o, ok := r.db[withdrawal.OrderID]; ok {
		if o.UID != withdrawal.UID {
			return ErrWithdrawalOwnedByAnotherUser
		}
		return ErrWithdrawalExists
	}
	r.db[withdrawal.OrderID] = withdrawal
	r.userWithdrawals[withdrawal.UID] = append(r.userWithdrawals[withdrawal.UID], withdrawal)
	return nil
}

func (r *InmemorWithdrawalRepository) GetUserWithdrawals(ctx context.Context, userID string) ([]*entity.Withdrawal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, ok := r.userWithdrawals[userID]
	if !ok {
		return []*entity.Withdrawal{}, nil
	}
	return orders, nil
}

func NewInmemoryWithdrawalRepository() *InmemorWithdrawalRepository {
	return &InmemorWithdrawalRepository{
		mu:              sync.RWMutex{},
		db:              make(map[string]*entity.Withdrawal, 100),
		userWithdrawals: make(map[string][]*entity.Withdrawal, 100),
	}
}
