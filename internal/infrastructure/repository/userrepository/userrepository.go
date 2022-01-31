package userrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type UserRepository interface {
	Get(ctx context.Context, login string) (entity.UserEntity, error)
	Add(ctx context.Context, entity entity.UserEntity) error
}
