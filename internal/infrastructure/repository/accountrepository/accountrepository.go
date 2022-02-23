package accountrepository

import (
	"context"
	"io"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type AccountRepository interface {
	AddAccount(ctx context.Context, account entity.Account) error
	GetAccount(ctx context.Context, userID string) (entity.Account, error)
	RefillAmount(ctx context.Context, userID string, diff float32) error
	WithdrawalAmount(ctx context.Context, userID string, diff float32) error
	io.Closer
}
