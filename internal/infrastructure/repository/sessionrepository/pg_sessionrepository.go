package sessionrepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgSessionRepository struct {
	db         *sql.DB
	statements map[queryType]*sql.Stmt
}

type queryType string

const (
	queryAddSession queryType = "addSession"
	queryGetSession queryType = "getSession"
)

var queries = map[queryType]string{
	queryAddSession: "insert into gophermart.sessions(sid, uid, created_at) values($1, $2, $3)",
	queryGetSession: "select sid, uid, created_at from gophermart.sessions where sid=$1",
}

func (p PgSessionRepository) AddSession(ctx context.Context, session *entity.Session) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[queryAddSession])
	_, err = stmt.ExecContext(ctx, session.SessionID, session.UID, session.CreatedAt)
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
	err := p.statements[queryGetSession].QueryRowContext(ctx, query, sessionID).Scan(&session.SessionID, &session.UID, &session.CreatedAt)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

func (p PgSessionRepository) Close() error {
	for name, stmt := range p.statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error close stmt %s: %w", name, err)
		}
	}
	return nil
}

func NewPgSessionRepository(db *sql.DB) (*PgSessionRepository, error) {
	statements := make(map[queryType]*sql.Stmt, len(queries))
	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("error prepare statement for %s: %w", name, err)
		}
		statements[name] = stmt
	}

	return &PgSessionRepository{db: db, statements: statements}, nil
}
