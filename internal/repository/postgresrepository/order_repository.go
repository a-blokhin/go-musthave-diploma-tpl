package postgresrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type orderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) repository.OrderRepository {
	return &orderRepository{pool: pool}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (id, number, user_id, status, accrual, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query, order.ID, order.Number, order.UserID, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {

		if pgErr, ok := err.(interface{ Code() string }); ok && pgErr.Code() == "23505" {
			return &model.OrderAlreadyExistsError{Number: order.Number}
		}
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *orderRepository) GetByNumber(ctx context.Context, number string) (*model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order
	err := r.pool.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user ID: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) GetOrdersByStatus(ctx context.Context, statuses []model.OrderStatus) ([]*model.Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE status = ANY($1)
		ORDER BY uploaded_at ASC
	`

	rows, err := r.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus, accrual float32) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE id = $3
	`

	_, err := r.pool.Exec(ctx, query, status, accrual, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// ExistsByNumber checks if an order with the given number exists
func (r *orderRepository) ExistsByNumber(ctx context.Context, number string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, number).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if order exists: %w", err)
	}

	return exists, nil
}
