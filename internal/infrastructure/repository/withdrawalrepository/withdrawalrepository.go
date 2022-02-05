package withdrawalrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type WithdrawalRepository interface {
	AddWithdrawal(ctx context.Context, order *entity.Withdrawal) error
	GetUserWithdrawals(ctx context.Context, userID string) ([]*entity.Withdrawal, error)
}
