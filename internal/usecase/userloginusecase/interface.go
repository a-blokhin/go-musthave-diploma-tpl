package userloginusecase

import (
	"context"

	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=UserRepository --output=./mocks --outpkg=mocks --with-expecter --filename=user_repository_mock.go
//go:generate mockery --name=PasswordService --output=./mocks --outpkg=mocks --with-expecter --filename=password_service_mock.go
//go:generate mockery --name=JWTService --output=./mocks --outpkg=mocks --with-expecter --filename=jwt_service_mock.go

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*model.User, error)
}

type PasswordService interface {
	CheckPassword(password, hash string) error
}

type JWTService interface {
	GenerateToken(userID uuid.UUID, login string) (string, error)
}
