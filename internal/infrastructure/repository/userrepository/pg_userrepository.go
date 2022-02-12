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

func (p PgUserRepository) AddUser(ctx context.Context, entity entity.UserEntity) error {
	query := "insert into gophermart.users(uid, login, password) values($1, $2, $3)"
	_, err := p.db.ExecContext(ctx, query, entity.UID, entity.Login, entity.Password)
	if err != nil {
		return err
	}
	return nil
}

func NewPgUserRepository(db *sql.DB) *PgUserRepository {
	return &PgUserRepository{db: db}
}
