package entity

import (
	"github.com/zaz600/go-musthave-diploma/internal/pkg/random"
)

// UserEntity пользователь системы накопления лояльности
type UserEntity struct {
	UID   string
	Login string
	// хэш пароля
	Password string
	// TODO дата регистрации, признак блокировки
}

type UserOption func(session *UserEntity)

func NewUserEntity(login string, password string, opts ...UserOption) UserEntity {
	user := UserEntity{
		UID:      random.UserID(),
		Login:    login,
		Password: password,
	}
	for _, opt := range opts {
		opt(&user)
	}
	return user
}
