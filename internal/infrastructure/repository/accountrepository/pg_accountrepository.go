package accountrepository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgAccountRepository struct {
	db         *sql.DB
	statements map[queryType]*sql.Stmt
}

type queryType string

const (
	queryInsertAccount     queryType = "insertAccount"
	querySelectAccount     queryType = "selectAccount"
	queryRefillAccount     queryType = "refillAccount"
	queryWithdrawalAccount queryType = "withdrawalAccount"
)

var queries = map[queryType]string{
	queryInsertAccount:     "insert into gophermart.accounts(uid, account_id, balance, withdrawals) values($1, $2, $3, $4)",
	querySelectAccount:     "select uid, account_id, balance, withdrawals from gophermart.accounts where uid=$1",
	queryRefillAccount:     "update gophermart.accounts set balance=balance+$1 where uid=$2",
	queryWithdrawalAccount: "update gophermart.accounts set balance=balance-$1, withdrawals=withdrawals+$1 where uid=$2",
}

func (p PgAccountRepository) AddAccount(ctx context.Context, account entity.Account) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	stmt := tx.Stmt(p.statements[queryInsertAccount])
	_, err = stmt.ExecContext(ctx, account.UID, account.AccountID, account.Balance, account.Withdrawals)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgAccountRepository) GetAccount(ctx context.Context, userID string) (entity.Account, error) {
	var account entity.Account
	err := p.statements[querySelectAccount].QueryRowContext(ctx, userID).Scan(&account.UID, &account.AccountID, &account.Balance, &account.Withdrawals)
	if err != nil {
		return entity.Account{}, ErrUserAccountNotFound
	}
	return account, nil
}

func (p PgAccountRepository) RefillAmount(ctx context.Context, userID string, amount float32) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	stmt := tx.Stmt(p.statements[queryRefillAccount])
	result, err := stmt.ExecContext(ctx, amount, userID)
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
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[queryWithdrawalAccount])
	result, err := stmt.ExecContext(ctx, amount, userID)
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

func (p PgAccountRepository) Close() error {
	for name, stmt := range p.statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error close stmt %s: %w", name, err)
		}
	}
	return nil
}

func NewPgAccountRepository(db *sql.DB) (*PgAccountRepository, error) {
	statements := make(map[queryType]*sql.Stmt, len(queries))
	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("error prepare statement for %s: %w", name, err)
		}
		statements[name] = stmt
	}

	return &PgAccountRepository{
		db:         db,
		statements: statements,
	}, nil
}
