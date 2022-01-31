package userservice

import (
	"context"
	"errors"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/utils/hasher"
)

type UserService interface {
	Register(ctx context.Context, login string, password string) (*entity.UserEntity, error)
	Login(ctx context.Context, login string, password string) (*entity.UserEntity, error)
	GetBalance(ctx context.Context) int
}

type Service struct {
	userRepository userrepository.UserRepository
}

func (s Service) Register(ctx context.Context, login string, password string) (*entity.UserEntity, error) {
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := entity.NewUserEntity(login, hashedPassword)
	err = s.userRepository.Add(ctx, user)
	if err != nil {
		if errors.Is(err, userrepository.ErrUserExists) {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return &user, nil
}

func (s Service) Login(ctx context.Context, login string, password string) (*entity.UserEntity, error) {
	user, err := s.userRepository.Get(ctx, login)
	if err != nil {
		if errors.Is(err, userrepository.ErrUserNotFound) {
			return nil, ErrAuth
		}
		return nil, err
	}

	if !hasher.CheckPasswordHash(password, user.Password) {
		return nil, ErrAuth
	}
	return &user, nil
}

func (s Service) GetBalance(ctx context.Context) int {
	return 0
}

func NewService(userRepo userrepository.UserRepository) *Service {
	return &Service{userRepository: userRepo}
}
