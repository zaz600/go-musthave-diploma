package orderrepository

import (
	"context"
	"io"
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type OrderRepository interface {
	AddOrder(ctx context.Context, order entity.Order) error
	SetOrderStatusAndAccrual(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error
	SetOrderNextRetryAt(ctx context.Context, orderID string, nextRetryAt time.Time) error
	GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetOrder(ctx context.Context, orderID string) (entity.Order, error)
	io.Closer
}
