package orderrepository

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemoryOrderRepository struct {
	mu         sync.RWMutex
	db         map[string]*entity.Order
	userOrders map[string][]*entity.Order
}

func (r *InmemoryOrderRepository) Add(_ context.Context, order *entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if o, ok := r.db[order.OrderID]; ok {
		if o.UID != order.UID {
			return ErrOrderOwnedByAnotherUser
		}
		return ErrOrderExists
	}
	r.db[order.OrderID] = order
	r.userOrders[order.UID] = append(r.userOrders[order.UID], order)
	return nil
}

func (r *InmemoryOrderRepository) GetUserOrders(_ context.Context, userID string) ([]*entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, ok := r.userOrders[userID]
	if !ok {
		return []*entity.Order{}, nil
	}
	return orders, nil
}

func NewInmemoryOrderRepository() *InmemoryOrderRepository {
	return &InmemoryOrderRepository{
		mu:         sync.RWMutex{},
		db:         make(map[string]*entity.Order, 100),
		userOrders: make(map[string][]*entity.Order, 100),
	}
}
