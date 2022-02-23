package userrepository

import (
	"context"
	"io"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type UserRepository interface {
	GetUser(ctx context.Context, login string) (entity.UserEntity, error)
	AddUser(ctx context.Context, entity entity.UserEntity) error
	io.Closer
}
