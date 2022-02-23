package orderrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgOrderRepository struct {
	db         *sql.DB
	statements map[queryType]*sql.Stmt
}

type queryType string

const (
	queryAddOrder            queryType = "AddOrder"
	querySetOrderStatus      queryType = "setOrderStatus"
	querySetOrderNextRetryAt queryType = "setOrderNextRetryAt"
	queryGetUserOrders       queryType = "GetUserOrders"
	queryGetOrder            queryType = "GetOrder"
)

var queries = map[queryType]string{
	queryAddOrder:            "insert into gophermart.orders(uid, order_id, status, accrual, retry_count) values($1, $2, $3, $4, $5)",
	querySetOrderStatus:      "update gophermart.orders set status=$1, accrual=$2 where order_id=$3",
	querySetOrderNextRetryAt: "update gophermart.orders set retry_count=retry_count+1 where order_id=$1",
	queryGetUserOrders:       "select uid, order_id, uploaded_at, status, accrual, retry_count from gophermart.orders where uid=$1",
	queryGetOrder:            "select uid, order_id, uploaded_at, status, accrual, retry_count from gophermart.orders where order_id=$1",
}

func (p PgOrderRepository) AddOrder(ctx context.Context, order entity.Order) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[queryAddOrder])
	_, err = stmt.ExecContext(ctx, order.UID, order.OrderID, order.Status, order.Accrual, order.RetryCount)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrOrderExists
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgOrderRepository) SetOrderStatusAndAccrual(ctx context.Context, orderID string, status entity.OrderStatus, accrual float32) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[querySetOrderStatus])
	_, err = stmt.ExecContext(ctx, status, accrual, orderID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgOrderRepository) SetOrderNextRetryAt(ctx context.Context, orderID string, nextRetryAt time.Time) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	stmt := tx.Stmt(p.statements[querySetOrderNextRetryAt])
	_, err = stmt.ExecContext(ctx, orderID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgOrderRepository) GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	rows, err := p.statements[queryGetUserOrders].QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(&order.UID, &order.OrderID, &order.UploadedAt, &order.Status, &order.Accrual, &order.RetryCount); err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return orders, nil
}

func (p PgOrderRepository) GetOrder(ctx context.Context, orderID string) (entity.Order, error) {
	var order entity.Order
	err := p.statements[queryGetOrder].QueryRowContext(ctx, orderID).Scan(&order.UID, &order.OrderID, &order.UploadedAt, &order.Status, &order.Accrual, &order.RetryCount)
	if err != nil {
		return order, ErrOrderNotFound
	}
	return order, nil
}

func (p PgOrderRepository) Close() error {
	for name, stmt := range p.statements {
		err := stmt.Close()
		if err != nil {
			return fmt.Errorf("error close stmt %s: %w", name, err)
		}
	}
	return nil
}

func NewPgOrderRepository(db *sql.DB) (*PgOrderRepository, error) {
	statements := make(map[queryType]*sql.Stmt, len(queries))
	for name, query := range queries {
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("error prepare statement for %s: %w", name, err)
		}
		statements[name] = stmt
	}

	return &PgOrderRepository{db: db, statements: statements}, nil
}
