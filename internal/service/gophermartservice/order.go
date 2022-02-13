package gophermartservice

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/utils/luhn"
)

func (s GophermartService) UploadOrder(ctx context.Context, userID string, orderID string) error {
	if ok := luhn.CheckLuhn(orderID); !ok {
		return ErrInvalidOrderFormat
	}

	order := *entity.NewOrder(userID, orderID)
	err := s.repo.OrderRepo.AddOrder(ctx, order)
	if err != nil {
		if errors.Is(err, orderrepository.ErrOrderExists) {
			order, err := s.repo.OrderRepo.GetOrder(ctx, orderID)
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

func (s GophermartService) GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	// TODO обработать ошибки и завернуть их
	orders, err := s.repo.OrderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].UploadedAt.UnixMilli() < orders[j].UploadedAt.UnixMilli()
	})
	return orders, nil
}
