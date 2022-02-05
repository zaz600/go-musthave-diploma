package orderrepository

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemoryOrderRepository struct {
	mu sync.RWMutex
	db map[string]*entity.Order
}

func (r *InmemoryOrderRepository) AddOrder(_ context.Context, order *entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if o, ok := r.db[order.OrderID]; ok {
		if o.UID != order.UID {
			return ErrOrderOwnedByAnotherUser
		}
		return ErrOrderExists
	}
	r.db[order.OrderID] = order
	return nil
}

func (r *InmemoryOrderRepository) UpdateOrder(_ context.Context, order *entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db[order.OrderID] = order
	return nil
}

func (r *InmemoryOrderRepository) GetUserOrders(_ context.Context, userID string) ([]*entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*entity.Order
	for _, order := range r.db {
		if order.UID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (r *InmemoryOrderRepository) GetOrder(_ context.Context, orderID string) (*entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.db[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func NewInmemoryOrderRepository() *InmemoryOrderRepository {
	return &InmemoryOrderRepository{
		mu: sync.RWMutex{},
		db: make(map[string]*entity.Order, 100),
	}
}
