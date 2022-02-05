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

func (r *InmemorWithdrawalRepository) AddWithdrawal(ctx context.Context, order *entity.Withdrawal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if o, ok := r.db[order.OrderID]; ok {
		if o.UID != order.UID {
			return ErrWithdrawalOwnedByAnotherUser
		}
		return ErrWithdrawalExists
	}
	r.db[order.OrderID] = order
	r.userWithdrawals[order.UID] = append(r.userWithdrawals[order.UID], order)
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
