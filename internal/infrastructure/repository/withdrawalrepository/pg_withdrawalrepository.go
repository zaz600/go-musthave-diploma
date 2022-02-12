package withdrawalrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgWithdrawalRepository struct {
	db *sql.DB
}

func (p PgWithdrawalRepository) AddWithdrawal(ctx context.Context, withdrawal entity.Withdrawal) error {
	query := "insert into gophermart.withdrawals(uid, order_id, processed_at, amount) values($1, $2, $3, $4)"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	_, err = tx.ExecContext(ctx, query, withdrawal.UID, withdrawal.OrderID, withdrawal.ProcessedAt, withdrawal.Sum)
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
	query := "select uid, order_id, processed_at, amount from gophermart.withdrawals where uid=$1"

	rows, err := p.db.QueryContext(ctx, query, userID)
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

func NewPgUserRepository(db *sql.DB) *PgWithdrawalRepository {
	return &PgWithdrawalRepository{db: db}
}
