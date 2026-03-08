package withdrawusecase

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name=BalanceRepository --output=./mocks --outpkg=mocks --with-expecter --filename=balance_repository_mock.go

type BalanceRepository interface {
	WithdrawWithRecord(ctx context.Context, userID uuid.UUID, orderNumber string, amount float32) error
}
