package postgresrepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type withdrawalRepository struct {
	pool *pgxpool.Pool
}

func NewWithdrawalRepository(pool *pgxpool.Pool) repository.WithdrawalRepository {
	return &withdrawalRepository{pool: pool}
}

func (r *withdrawalRepository) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (id, order_number, user_id, sum, processed_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query, withdrawal.ID, withdrawal.OrderNumber, withdrawal.UserID, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {

		if pgErr, ok := err.(interface{ Code() string }); ok && pgErr.Code() == "23505" {
			return fmt.Errorf("withdrawal with order number %s already exists", withdrawal.OrderNumber)
		}
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return nil
}

func (r *withdrawalRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error) {
	query := `
		SELECT id, order_number, user_id, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals by user ID: %w", err)
	}
	defer rows.Close()

	var withdrawals []*model.Withdrawal
	for rows.Next() {
		var withdrawal model.Withdrawal
		err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.OrderNumber,
			&withdrawal.UserID,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (r *withdrawalRepository) ExistsByOrderNumber(ctx context.Context, orderNumber string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM withdrawals WHERE order_number = $1)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, orderNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if withdrawal exists: %w", err)
	}

	return exists, nil
}
