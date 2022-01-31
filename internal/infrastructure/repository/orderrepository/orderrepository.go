package orderrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type OrderRepository interface {
	Add(ctx context.Context, order *entity.Order) error
	GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error)
}
