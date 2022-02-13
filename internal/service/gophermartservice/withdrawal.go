package gophermartservice

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

func (s GophermartService) UploadWithdrawal(ctx context.Context, userID string, orderID string, sum float32) error {
	withdrawal := entity.NewWithdrawal(userID, orderID, sum)
	return s.repo.WithdrawalRepo.AddWithdrawal(ctx, *withdrawal)
}

func (s GophermartService) GetUserWithdrawals(ctx context.Context, userID string) ([]entity.Withdrawal, error) {
	return s.repo.WithdrawalRepo.GetUserWithdrawals(ctx, userID)
}
