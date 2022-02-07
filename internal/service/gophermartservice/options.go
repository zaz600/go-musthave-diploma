package gophermartservice

import (
	"github.com/rs/zerolog/log"
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

const (
	Memory = iota
	PG
)

type Option func(*GophermartService)

func WithStorage(storageType StorageType) Option {
	return func(s *GophermartService) {
		switch storageType {
		case Memory:
			s.userService = userservice.NewService(userrepository.NewInmemoryUserRepository())
			s.sessionService = sessionservice.NewService(sessionrepository.NewInmemorySessionRepository())
			s.OrderService = orderservice.NewService(orderrepository.NewInmemoryOrderRepository())
			s.WithdrawalService = withdrawalservice.NewService(withdrawalrepository.NewInmemoryWithdrawalRepository())
		default:
			log.Panic().Msg("unsupported storage type")
		}
	}
}
