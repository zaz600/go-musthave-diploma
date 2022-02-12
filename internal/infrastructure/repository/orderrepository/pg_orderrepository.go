package orderrepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/zaz600/go-musthave-diploma/internal/entity"
)

type PgOrderRepository struct {
	db *sql.DB
}

func (p PgOrderRepository) AddOrder(ctx context.Context, order entity.Order) error {
	query := "insert into gophermart.orders(uid, order_id, status, accrual, retry_count) values($1, $2, $3, $4, $5)"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	_, err = tx.ExecContext(ctx, query, order.UID, order.OrderID, order.Status, order.Accrual, order.RetryCount)
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
	query := "update gophermart.orders set status=$1, accrual=$2 where order_id=$3"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, query, status, accrual, orderID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgOrderRepository) SetOrderNextRetryAt(ctx context.Context, orderID string, nextRetryAt time.Time) error {
	query := "update gophermart.orders set retry_count=retry_count+1 where order_id=$1"
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx, query, orderID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p PgOrderRepository) GetUserOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	query := "select uid, order_id, uploaded_at, status, accrual, retry_count from gophermart.orders where uid=$1"

	rows, err := p.db.QueryContext(ctx, query, userID)
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

func (p PgOrderRepository) GetUserAccrual(ctx context.Context, userID string) (float32, error) {
	query := "select coalesce(sum(accrual), 0) as total from gophermart.orders where uid=$1"

	var accrual float32
	err := p.db.QueryRowContext(ctx, query, userID).Scan(&accrual)
	if err != nil {
		return 0, err
	}
	return accrual, nil
}

func (p PgOrderRepository) GetOrder(ctx context.Context, orderID string) (entity.Order, error) {
	query := "select uid, order_id, uploaded_at, status, accrual, retry_count from gophermart.orders where order_id=$1"

	var order entity.Order
	err := p.db.QueryRowContext(ctx, query, orderID).Scan(&order.UID, &order.OrderID, &order.UploadedAt, &order.Status, &order.Accrual, &order.RetryCount)
	if err != nil {
		return order, ErrOrderNotFound
	}
	return order, nil
}

func NewPgOrderRepository(db *sql.DB) *PgOrderRepository {
	return &PgOrderRepository{db: db}
}
