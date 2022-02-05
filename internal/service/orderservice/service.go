package orderservice

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/utils/luhn"
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID string, orderID string) error
	GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error)
	SetOrderStatus(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error
}

type Service struct {
	orderRepository orderrepository.OrderRepository
}

func (s Service) UploadOrder(ctx context.Context, userID string, orderID string) error {
	if ok := luhn.CheckLuhn(orderID); !ok {
		return ErrInvalidOrderFormat
	}

	order := entity.NewOrder(userID, orderID)
	err := s.orderRepository.AddOrder(ctx, order)
	if err != nil {
		if errors.Is(err, orderrepository.ErrOrderExists) {
			return ErrOrderExists
		}
		if errors.Is(err, orderrepository.ErrOrderOwnedByAnotherUser) {
			return ErrOrderOwnedByAnotherUser
		}
		return fmt.Errorf("error upload order %s, uid=%s: %w", orderID, userID, err)
	}
	return nil
}

func (s Service) GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error) {
	// TODO обработать ошибки и завернуть их
	orders, err := s.orderRepository.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].UploadedAt < orders[j].UploadedAt
	})
	return orders, nil
}

func (s Service) SetOrderStatus(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error {
	order, err := s.orderRepository.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}
	order.Status = status
	order.Accrual = accrual
	return s.orderRepository.UpdateOrder(ctx, order)
}

func NewService(orderRepository orderrepository.OrderRepository) *Service {
	return &Service{
		orderRepository: orderRepository,
	}
}
