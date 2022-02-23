package withdrawalrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgWithdrawalRepository struct {
	db         *sql.DB
	statements map[queryType]*sql.Stmt
}

type queryType string

const (
	queryAddWithdrawal      queryType = "AddWithdrawal"
	queryGetUserWithdrawals queryType = "GetUserWithdrawals"
)

var queries = map[queryType]string{
	queryAddWithdrawal:      "insert into gophermart.withdrawals(uid, order_id, processed_at, amount) values($1, $2, $3, $4)",
	queryGetUserWithdrawals: "select uid, order_id, processed_at, amount from gophermart.withdrawals where uid=$1",
}

func (p PgWithdrawalRepository) AddWithdrawal(ctx context.Context, withdrawal entity.Withdrawal) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[queryAddWithdrawal])
	_, err = stmt.ExecContext(ctx, withdrawal.UID, withdrawal.OrderID, withdrawal.ProcessedAt, withdrawal.Sum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrWithdrawalExists
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgWithdrawalRepository) GetUserWithdrawals(ctx context.Context, userID string) ([]entity.Withdrawal, error) {
	rows, err := p.statements[queryGetUserWithdrawals].QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var withdrawals []entity.Withdrawal
	for rows.Next() {
		var withdrawal entity.Withdrawal
		if err := rows.Scan(&withdrawal.UID, &withdrawal.OrderID, &withdrawal.ProcessedAt, &withdrawal.Sum); err != nil {
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return withdrawals, nil
}

func (p PgWithdrawalRepository) Close() error {
	for name, stmt := range p.statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error close stmt %s: %w", name, err)
		}
	}
	return nil
}

func NewPgUserRepository(db *sql.DB) (*PgWithdrawalRepository, error) {
	statements := make(map[queryType]*sql.Stmt, len(queries))
	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("error prepare statement for %s: %w", name, err)
		}
		statements[name] = stmt
	}

	return &PgWithdrawalRepository{db: db, statements: statements}, nil
}
