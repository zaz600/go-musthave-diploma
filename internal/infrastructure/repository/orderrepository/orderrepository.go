package orderrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type OrderRepository interface {
	AddOrder(ctx context.Context, order *entity.Order) error
	UpdateOrder(ctx context.Context, order *entity.Order) error // TODO разбить на дискретные методы апдейта отдельных полей
	GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error)
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
}
