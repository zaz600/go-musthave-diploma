package sessionrepository

import (
	"context"
	"io"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type SessionRepository interface {
	AddSession(ctx context.Context, session *entity.Session) error
	GetSession(ctx context.Context, sessionID string) (*entity.Session, error)
	io.Closer
}
