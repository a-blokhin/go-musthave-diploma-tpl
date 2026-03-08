package postgresrepository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, login, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Login, user.PasswordHash, user.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(interface{ Code() string }); ok && pgErr.Code() == "23505" {
			return &model.UserAlreadyExistsError{Login: user.Login}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}
