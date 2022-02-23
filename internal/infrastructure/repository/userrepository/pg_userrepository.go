package userrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgUserRepository struct {
	db         *sql.DB
	statements map[queryType]*sql.Stmt
}

type queryType string

const (
	queryGetUser queryType = "getUser"
	queryAddUser queryType = "addUser"
)

var queries = map[queryType]string{
	queryGetUser: "select uid, login, password from gophermart.users where login=$1",
	queryAddUser: "insert into gophermart.users(uid, login, password) values($1, $2, $3)",
}

func (p PgUserRepository) GetUser(ctx context.Context, login string) (entity.UserEntity, error) {
	var user entity.UserEntity
	err := p.statements[queryGetUser].QueryRowContext(ctx, login).Scan(&user.UID, &user.Login, &user.Password)
	if err != nil {
		return user, ErrUserNotFound
	}
	return user, nil
}

func (p PgUserRepository) AddUser(ctx context.Context, userEntity entity.UserEntity) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[queryAddUser])
	_, err = stmt.ExecContext(ctx, userEntity.UID, userEntity.Login, userEntity.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserExists
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgUserRepository) Close() error {
	for name, stmt := range p.statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error close stmt %s: %w", name, err)
		}
	}
	return nil
}

func NewPgUserRepository(db *sql.DB) (*PgUserRepository, error) {
	statements := make(map[queryType]*sql.Stmt, len(queries))
	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("error prepare statement for %s: %w", name, err)
		}
		statements[name] = stmt
	}

	return &PgUserRepository{db: db, statements: statements}, nil
}
