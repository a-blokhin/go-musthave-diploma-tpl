package userregistrationusecase

import (
	"context"

	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=UserRepository --output=./mocks --outpkg=mocks --with-expecter --filename=user_repository_mock.go
//go:generate mockery --name=PasswordService --output=./mocks --outpkg=mocks --with-expecter --filename=password_service_mock.go
//go:generate mockery --name=JWTService --output=./mocks --outpkg=mocks --with-expecter --filename=jwt_service_mock.go
//go:generate mockery --name=BalanceRepository --output=./mocks --outpkg=mocks --with-expecter --filename=balance_repository_mock.go

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

type PasswordService interface {
	HashPassword(password string) (string, error)
}

type JWTService interface {
	GenerateToken(userID uuid.UUID, login string) (string, error)
}

type BalanceRepository interface {
	Create(ctx context.Context, balance *model.Balance) error
}
