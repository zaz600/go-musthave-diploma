package gophermartservice

import "context"

func (s GophermartService) GetUserBalance(ctx context.Context, userID string) (float32, float32, error) {
	balance, err := s.repo.OrderRepo.GetUserAccrual(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	var withdrawalsSum float32
	withdrawals, err := s.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	for _, withdrawal := range withdrawals {
		withdrawalsSum += withdrawal.Sum
	}
	return balance - withdrawalsSum, withdrawalsSum, nil
}
