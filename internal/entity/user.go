package entity

import (
	"github.com/zaz600/go-musthave-diploma/internal/utils/random"
)

type UserEntity struct {
	UID   string
	Login string
	// хэш пароля
	Password string
}

func NewUserEntity(login string, password string) UserEntity {
	return UserEntity{
		UID:      random.UserID(),
		Login:    login,
		Password: password,
	}
}
