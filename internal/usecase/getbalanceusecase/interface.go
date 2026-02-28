package getbalanceusecase

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=BalanceRepository --output=./mocks --outpkg=mocks --with-expecter --filename=balance_repository_mock.go

type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (model.User, error)
}
