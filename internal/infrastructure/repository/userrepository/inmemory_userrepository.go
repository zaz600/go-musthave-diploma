package userrepository

import (
	"context"
	"sync"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type InmemoryUserRepository struct {
	mu sync.RWMutex
	db map[string]entity.UserEntity
}

func NewInmemoryUserRepository() *InmemoryUserRepository {
	return &InmemoryUserRepository{
		mu: sync.RWMutex{},
		db: make(map[string]entity.UserEntity, 100),
	}
}

func (r *InmemoryUserRepository) Get(_ context.Context, login string) (entity.UserEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userEntity, ok := r.db[login]; ok {
		return userEntity, nil
	}
	return entity.UserEntity{}, ErrUserNotFound
}

func (r *InmemoryUserRepository) Add(_ context.Context, userEntity entity.UserEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[userEntity.Login]; ok {
		return ErrUserExists
	}
	r.db[userEntity.Login] = userEntity
	return nil
}
