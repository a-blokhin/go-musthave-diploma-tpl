package withdrawusecase

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=BalanceRepository --output=./mocks --outpkg=mocks --with-expecter --filename=balance_repository_mock.go
//go:generate mockery --name=WithdrawalRepository --output=./mocks --outpkg=mocks --with-expecter --filename=withdrawal_repository_mock.go

type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	Withdraw(ctx context.Context, userID uuid.UUID, amount float32) error
}

type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *model.Withdrawal) error
	ExistsByOrderNumber(ctx context.Context, orderNumber string) (bool, error)
}
