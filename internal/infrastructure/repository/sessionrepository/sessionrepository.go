package sessionrepository

import (
	"context"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type SessionRepository interface {
	AddSession(ctx context.Context, session *entity.Session) error
	DelSession(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*entity.Session, error)
}
