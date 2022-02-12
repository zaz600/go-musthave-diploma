package gophermartservice

import (
	"database/sql"
	"time"

	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/withdrawalrepository"
	"github.com/zaz600/go-musthave-diploma/internal/service/orderservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/sessionservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/userservice"
	"github.com/zaz600/go-musthave-diploma/internal/service/withdrawalservice"
)

type StorageType int

type Option func(*GophermartService)

func WithMemoryStorage() Option {
	return func(s *GophermartService) {
		s.userService = userservice.NewService(userrepository.NewInmemoryUserRepository())
		s.sessionService = sessionservice.NewService(sessionrepository.NewInmemorySessionRepository())
		s.OrderService = orderservice.NewService(orderrepository.NewInmemoryOrderRepository())
		s.WithdrawalService = withdrawalservice.NewService(withdrawalrepository.NewInmemoryWithdrawalRepository())
	}
}

func WithPgStorage(db *sql.DB) Option {
	return func(s *GophermartService) {
		s.userService = userservice.NewService(userrepository.NewPgUserRepository(db))
		s.sessionService = sessionservice.NewService(sessionrepository.NewPgSessionRepository(db))
		s.OrderService = orderservice.NewService(orderrepository.NewInmemoryOrderRepository())
		s.WithdrawalService = withdrawalservice.NewService(withdrawalrepository.NewInmemoryWithdrawalRepository())
	}
}

func WithAccrualRetryInterval(interval time.Duration) Option {
	return func(s *GophermartService) {
		s.accrualRetryInterval = interval
	}
}
