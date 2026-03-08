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
