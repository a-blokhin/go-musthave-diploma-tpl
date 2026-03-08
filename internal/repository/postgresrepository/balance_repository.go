package postgresrepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type balanceRepository struct {
	pool *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) repository.BalanceRepository {
	return &balanceRepository{pool: pool}
}

func (r *balanceRepository) Create(ctx context.Context, balance *model.Balance) error {
	query := `
		INSERT INTO user_balances (user_id, current_balance, withdrawn_balance, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, balance.UserID, balance.CurrentBalance, balance.WithdrawnBalance, balance.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create balance: %w", err)
	}

	return nil
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	query := `
		SELECT user_id, current_balance, withdrawn_balance, updated_at
		FROM user_balances
		WHERE user_id = $1
	`

	var balance model.Balance
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.CurrentBalance,
		&balance.WithdrawnBalance,
		&balance.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get balance by user ID: %w", err)
	}

	return &balance, nil
}

func (r *balanceRepository) Update(ctx context.Context, balance *model.Balance) error {
	query := `
		UPDATE user_balances
		SET current_balance = $1, withdrawn_balance = $2, updated_at = $3
		WHERE user_id = $4
	`

	_, err := r.pool.Exec(ctx, query, balance.CurrentBalance, balance.WithdrawnBalance, balance.UpdatedAt, balance.UserID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

func (r *balanceRepository) AddAccrual(ctx context.Context, userID uuid.UUID, amount float32) error {
	query := `
		UPDATE user_balances
		SET current_balance = current_balance + $1, updated_at = NOW()
		WHERE user_id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to add accrual: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("balance not found for user")
	}

	return nil
}

func (r *balanceRepository) WithdrawWithRecord(ctx context.Context, userID uuid.UUID, orderNumber string, amount float32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentBalance float32
	balanceQuery := `SELECT current_balance FROM user_balances WHERE user_id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, balanceQuery, userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if currentBalance < amount {
		return model.ErrInsufficientBalance
	}

	updateBalanceQuery := `
		UPDATE user_balances
		SET current_balance = current_balance - $1,
			withdrawn_balance = withdrawn_balance + $1,
			updated_at = NOW()
		WHERE user_id = $2
	`
	_, err = tx.Exec(ctx, updateBalanceQuery, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	withdrawalQuery := `
		INSERT INTO withdrawals (id, order_number, user_id, sum, processed_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
	`
	_, err = tx.Exec(ctx, withdrawalQuery, orderNumber, userID, amount)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
