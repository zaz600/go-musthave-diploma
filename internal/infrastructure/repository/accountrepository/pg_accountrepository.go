package accountrepository

import (
	"context"
	"database/sql"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgAccountRepository struct {
	db *sql.DB
}

func (p PgAccountRepository) AddAccount(ctx context.Context, account entity.Account) (err error) {
	query := "insert into gophermart.accounts(uid, account_id, balance, withdrawals) values($1, $2, $3, $4)"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, query, account.UID, account.AccountID, account.Balance, account.Withdrawals)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgAccountRepository) GetAccount(ctx context.Context, userID string) (entity.Account, error) {
	query := "select uid, account_id, balance, withdrawals from gophermart.accounts where uid=$1"

	var account entity.Account
	err := p.db.QueryRowContext(ctx, query, userID).Scan(&account.UID, &account.AccountID, &account.Balance, &account.Withdrawals)
	if err != nil {
		return entity.Account{}, ErrUserAccountNotFound
	}
	return account, nil
}

func (p PgAccountRepository) RefillAmount(ctx context.Context, userID string, amount float32) error {
	query := "update gophermart.accounts set balance=balance+$1 where uid=$2"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	result, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserAccountNotFound
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgAccountRepository) WithdrawalAmount(ctx context.Context, userID string, amount float32) error {
	query := "update gophermart.accounts set balance=balance-$1, withdrawals=withdrawals+$1 where uid=$2"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	result, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserAccountNotFound
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func NewPgAccountRepository(db *sql.DB) *PgAccountRepository {
	return &PgAccountRepository{db: db}
}
