package gophermartservice

import "context"

func (s GophermartService) GetUserBalance(ctx context.Context, userID string) (float32, float32, error) {
	account, err := s.repo.AccountRepo.GetAccount(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return account.Balance, account.Withdrawals, nil
}
