package withdrawalservice

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
	"github.com/zaz600/go-musthave-diploma/internal/infrastructure/repository/withdrawalrepository"
)

type WithdrawalService interface {
	UploadWithdrawal(ctx context.Context, userID string, orderID string, sum float32) error
	GetUserWithdrawals(ctx context.Context, userID string) ([]entity.Withdrawal, error)
}

type Service struct {
	withdrawalRepository withdrawalrepository.WithdrawalRepository
}

func (s *Service) UploadWithdrawal(ctx context.Context, userID string, orderID string, sum float32) error {
	withdrawal := entity.NewWithdrawal(userID, orderID, sum)
	return s.withdrawalRepository.AddWithdrawal(ctx, *withdrawal)
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID string) ([]entity.Withdrawal, error) {
	return s.withdrawalRepository.GetUserWithdrawals(ctx, userID)
}

func NewService(withdrawalRepo withdrawalrepository.WithdrawalRepository) *Service {
	return &Service{
		withdrawalRepository: withdrawalRepo,
	}
}
