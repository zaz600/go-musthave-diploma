package orderrepository

import (
	"context"
	"sync"
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemoryOrderRepository struct {
	mu sync.RWMutex
	db map[string]entity.Order
}

func (r *InmemoryOrderRepository) AddOrder(_ context.Context, order entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[order.OrderID]; ok {
		return ErrOrderExists
	}
	r.db[order.OrderID] = order
	return nil
}

func (r *InmemoryOrderRepository) UpdateOrder(_ context.Context, order entity.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db[order.OrderID] = order
	return nil
}

func (r *InmemoryOrderRepository) SetOrderStatusAndAccrual(_ context.Context, orderID string, status entity.OrderStatus, accrual float32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order, ok := r.db[orderID]; ok {
		order.Status = status
		order.Accrual = accrual
		r.db[orderID] = order
		return nil
	}
	return ErrOrderNotFound
}

func (r *InmemoryOrderRepository) SetOrderNextRetryAt(_ context.Context, orderID string, nextRetryAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order, ok := r.db[orderID]; ok {
		order.RetryCount++
		r.db[orderID] = order
		return nil
	}
	return ErrOrderNotFound
}

func (r *InmemoryOrderRepository) GetUserOrders(_ context.Context, userID string) ([]entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []entity.Order
	for _, order := range r.db {
		if order.UID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func (r *InmemoryOrderRepository) GetUserAccrual(ctx context.Context, userID string) (float32, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, err := r.GetUserOrders(ctx, userID)
	if err != nil {
		return 0, err
	}
	var sum float32
	for _, order := range orders {
		if order.Status == entity.OrderStatusPROCESSED {
			sum += order.Accrual
		}
	}
	return sum, nil
}

func (r *InmemoryOrderRepository) GetOrder(_ context.Context, orderID string) (entity.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, ok := r.db[orderID]
	if !ok {
		return entity.Order{}, ErrOrderNotFound
	}
	return order, nil
}

func NewInmemoryOrderRepository() *InmemoryOrderRepository {
	return &InmemoryOrderRepository{
		mu: sync.RWMutex{},
		db: make(map[string]entity.Order, 100),
	}
}
