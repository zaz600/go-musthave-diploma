package gophermartservice

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

func (s GophermartService) UploadWithdrawal(ctx context.Context, userID string, orderID string, sum float32) error {
	// TODO баланс изменять надо в одной транзакции с регистрацией списания

	withdrawal := entity.NewWithdrawal(userID, orderID, sum)
	err := s.repo.WithdrawalRepo.AddWithdrawal(ctx, withdrawal)
	if err != nil {
		return err
	}
	return s.repo.AccountRepo.WithdrawalAmount(ctx, userID, sum)
}

func (s GophermartService) GetUserWithdrawals(ctx context.Context, userID string) ([]entity.Withdrawal, error) {
	return s.repo.WithdrawalRepo.GetUserWithdrawals(ctx, userID)
}
