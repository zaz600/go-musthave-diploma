package sessionrepository

import (
	"context"
	"database/sql"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgSessionRepository struct {
	db *sql.DB
}

func (p PgSessionRepository) AddSession(ctx context.Context, session *entity.Session) error {
	query := "insert into gophermart.sessions(sid, uid, created_at) values($1, $2, $3)"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, query, session.SessionID, session.UID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgSessionRepository) GetSession(ctx context.Context, sessionID string) (*entity.Session, error) {
	query := "select sid, uid, created_at from gophermart.sessions where sid=$1"

	var session entity.Session
	err := p.db.QueryRowContext(ctx, query, sessionID).Scan(&session.SessionID, &session.UID, &session.CreatedAt)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

func NewPgSessionRepository(db *sql.DB) *PgSessionRepository {
	return &PgSessionRepository{db: db}
}
