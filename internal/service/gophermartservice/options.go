package gophermartservice

import (
	"database/sql"
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/accountrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/withdrawalrepository"
)

type StorageType int

type Option func(*GophermartService) error

func WithMemoryStorage() Option {
	return func(s *GophermartService) error {
		repo := repository.RepoRegistry{
			UserRepo:       userrepository.NewInmemoryUserRepository(),
			SessionRepo:    sessionrepository.NewInmemorySessionRepository(),
			OrderRepo:      orderrepository.NewInmemoryOrderRepository(),
			WithdrawalRepo: withdrawalrepository.NewInmemoryWithdrawalRepository(),
			AccountRepo:    accountrepository.NewInmemoryAccountRepository(),
		}
		s.repo = repo
		return nil
	}
}

func WithPgStorage(db *sql.DB) Option {
	return func(s *GophermartService) error {
		accrualRepo, err := accountrepository.NewPgAccountRepository(db)
		if err != nil {
			return err
		}
		orderRepo, err := orderrepository.NewPgOrderRepository(db)
		if err != nil {
			return err
		}
		sessionRepo, err := sessionrepository.NewPgSessionRepository(db)
		if err != nil {
			return err
		}

		repo := repository.RepoRegistry{
			UserRepo:       userrepository.NewPgUserRepository(db),
			SessionRepo:    sessionRepo,
			OrderRepo:      orderRepo,
			WithdrawalRepo: withdrawalrepository.NewPgUserRepository(db),
			AccountRepo:    accrualRepo,
		}
		s.repo = repo
		return nil
	}
}

func WithAccrualRetryInterval(interval time.Duration) Option {
	return func(s *GophermartService) error {
		s.accrualRetryInterval = interval
		return nil
	}
}
