package sessionrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type SessionRepository interface {
	Add(ctx context.Context, session *entity.Session) error
	Del(ctx context.Context, sessionID string) error
	Get(ctx context.Context, sessionID string) (*entity.Session, error)
}
