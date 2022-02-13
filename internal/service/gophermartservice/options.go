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

type Option func(*GophermartService)

func WithMemoryStorage() Option {
	return func(s *GophermartService) {
		repo := repository.RepoRegistry{
			UserRepo:       userrepository.NewInmemoryUserRepository(),
			SessionRepo:    sessionrepository.NewInmemorySessionRepository(),
			OrderRepo:      orderrepository.NewInmemoryOrderRepository(),
			WithdrawalRepo: withdrawalrepository.NewInmemoryWithdrawalRepository(),
			AccountRepo:    accountrepository.NewInmemoryAccountRepository(),
		}
		s.repo = repo
	}
}

func WithPgStorage(db *sql.DB) Option {
	return func(s *GophermartService) {
		repo := repository.RepoRegistry{
			UserRepo:       userrepository.NewPgUserRepository(db),
			SessionRepo:    sessionrepository.NewPgSessionRepository(db),
			OrderRepo:      orderrepository.NewPgOrderRepository(db),
			WithdrawalRepo: withdrawalrepository.NewPgUserRepository(db),
			AccountRepo:    accountrepository.NewPgAccountRepository(db),
		}
		s.repo = repo
	}
}

func WithAccrualRetryInterval(interval time.Duration) Option {
	return func(s *GophermartService) {
		s.accrualRetryInterval = interval
	}
}
