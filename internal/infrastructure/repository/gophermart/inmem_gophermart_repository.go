package gophermart

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InMemGophermartRepository struct {
	muUsers sync.RWMutex
	users   map[string]entity.UserEntity
}

func (r *InMemGophermartRepository) GetUser(ctx context.Context, login string) (entity.UserEntity, error) {
	r.muUsers.RLock()
	defer r.muUsers.RUnlock()

	if userEntity, ok := r.users[login]; ok {
		return userEntity, nil
	}
	return entity.UserEntity{}, ErrUserNotFound
}

func (r *InMemGophermartRepository) AddUser(ctx context.Context, entity entity.UserEntity) error {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) AddSession(ctx context.Context, session *entity.Session) error {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) GetSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) AddOrder(ctx context.Context, order *entity.Order) error {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) GetUserOrders(ctx context.Context, userID string) ([]*entity.Order, error) {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) AddWithdrawal(ctx context.Context, order *entity.Withdrawal) error {
	// TODO implement me
	panic("implement me")
}

func (r *InMemGophermartRepository) GetUserWithdrawals(ctx context.Context, userID string) ([]*entity.Withdrawal, error) {
	// TODO implement me
	panic("implement me")
}
