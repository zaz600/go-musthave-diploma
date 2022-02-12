package orderservice

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/utils/luhn"
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID string, orderID string) error
	GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetUserAccrual(ctx context.Context, userID string) (float32, error)
	SetOrderStatus(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error
	ReScheduleOrderProcessingTask(ctx context.Context, orderID string, next time.Time) error
	GetOrder(ctx context.Context, orderID string) (entity.Order, error)
}

type Service struct {
	orderRepository orderrepository.OrderRepository
}

func (s Service) UploadOrder(ctx context.Context, userID string, orderID string) error {
	if ok := luhn.CheckLuhn(orderID); !ok {
		return ErrInvalidOrderFormat
	}

	order := *entity.NewOrder(userID, orderID)
	err := s.orderRepository.AddOrder(ctx, order)
	if err != nil {
		if errors.Is(err, orderrepository.ErrOrderExists) {
			order, err := s.orderRepository.GetOrder(ctx, orderID)
			if err != nil {
				return fmt.Errorf("error upload order %s, uid=%s: %w", orderID, userID, err)
			}
			if order.UID != userID {
				return ErrOrderOwnedByAnotherUser
			}
			return ErrOrderExists
		}
		return fmt.Errorf("error upload order %s, uid=%s: %w", orderID, userID, err)
	}
	return nil
}

func (s Service) GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	// TODO обработать ошибки и завернуть их
	orders, err := s.orderRepository.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].UploadedAt.UnixMilli() < orders[j].UploadedAt.UnixMilli()
	})
	return orders, nil
}

func (s Service) GetUserAccrual(ctx context.Context, userID string) (float32, error) {
	return s.orderRepository.GetUserAccrual(ctx, userID)
}

func (s Service) SetOrderStatus(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error {
	return s.orderRepository.SetOrderStatusAndAccrual(ctx, orderID, status, accrual)
}

func (s Service) GetOrder(ctx context.Context, orderID string) (entity.Order, error) {
	return s.orderRepository.GetOrder(ctx, orderID)
}

func (s Service) ReScheduleOrderProcessingTask(ctx context.Context, orderID string, next time.Time) error {
	return s.orderRepository.SetOrderNextRetryAt(ctx, orderID, next)
}

func NewService(orderRepository orderrepository.OrderRepository) *Service {
	return &Service{
		orderRepository: orderRepository,
	}
}
