package gophermart

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type GophermartRepository interface {
	GetUser(ctx context.Context, login string) (entity.UserEntity, error)
	AddUser(ctx context.Context, entity entity.UserEntity) error

	AddSession(ctx context.Context, session *entity.Session) error
	GetSession(ctx context.Context, sessionID string) (*entity.Session, error)

	AddOrder(ctx context.Context, order *entity.Order) error
	GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error)

	AddWithdrawal(ctx context.Context, order *entity.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, userID string) ([]*entity.Withdrawal, error)
}
