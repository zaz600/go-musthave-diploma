package userrepository

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgUserRepository struct {
	db *sql.DB
}

func (p PgUserRepository) GetUser(ctx context.Context, login string) (entity.UserEntity, error) {
	query := "select uid, login, password from gophermart.users where login=$1"

	var user entity.UserEntity
	err := p.db.QueryRowContext(ctx, query, login).Scan(&user.UID, &user.Login, &user.Password)
	if err != nil {
		return user, ErrUserNotFound
	}
	return user, nil
}

func (p PgUserRepository) AddUser(ctx context.Context, userEntity entity.UserEntity) error {
	query := "insert into gophermart.users(uid, login, password) values($1, $2, $3)"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, query, userEntity.UID, userEntity.Login, userEntity.Password)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func NewPgUserRepository(db *sql.DB) *PgUserRepository {
	return &PgUserRepository{db: db}
}
