package gophermartservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/pkg/hasher"
)

func (s GophermartService) RegisterUser(ctx context.Context, login string, password string) (*entity.Session, error) {
	// TODO создавать сессию и открывать счет надо в одной транзакции с регистрацией

	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := entity.NewUserEntity(login, hashedPassword)
	err = s.repo.UserRepo.AddUser(ctx, user)
	if err != nil {
		if errors.Is(err, userrepository.ErrUserExists) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	account := entity.NewAccount(user.UID)
	err = s.repo.AccountRepo.AddAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	return s.createSession(ctx, user)
}

func (s GophermartService) LoginUser(ctx context.Context, login string, password string) (*entity.Session, error) {
	user, err := s.repo.UserRepo.GetUser(ctx, login)
	if err != nil {
		if errors.Is(err, userrepository.ErrUserNotFound) {
			return nil, ErrAuth
		}
		return nil, err
	}

	if !hasher.CheckPasswordHash(password, user.Password) {
		return nil, ErrAuth
	}

	return s.createSession(ctx, user)
}

func (s GophermartService) createSession(ctx context.Context, user entity.UserEntity) (*entity.Session, error) {
	session := entity.NewRandomSession(user.UID)
	if err := s.repo.SessionRepo.AddSession(ctx, session); err != nil {
		return nil, fmt.Errorf("error creating user session: %w", err)
	}
	return session, nil
}
