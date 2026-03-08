package getwithdrawalsusecase

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=WithdrawalRepository --output=./mocks --outpkg=mocks --with-expecter --filename=withdrawal_repository_mock.go

type WithdrawalRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Withdrawal, error)
}
