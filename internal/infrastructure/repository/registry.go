package repository

import (
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/accountrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/orderrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/sessionrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/userrepository"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/withdrawalrepository"
)

type RepoRegistry struct {
	UserRepo       userrepository.UserRepository
	SessionRepo    sessionrepository.SessionRepository
	OrderRepo      orderrepository.OrderRepository
	WithdrawalRepo withdrawalrepository.WithdrawalRepository
	AccountRepo    accountrepository.AccountRepository
}

func (r *RepoRegistry) Close() {
	_ = r.AccountRepo.Close()
	_ = r.OrderRepo.Close()
	_ = r.SessionRepo.Close()
}
