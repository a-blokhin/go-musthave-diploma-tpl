package repository

import (
	"context"

	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=OrderRepository --output=../service/mocks --outpkg=mocks --with-expecter --filename=order_repository_mock.go
//go:generate mockery --name=BalanceRepository --output=../service/mocks --outpkg=mocks --with-expecter --filename=balance_repository_mock.go
//go:generate mockery --name=WithdrawalRepository --output=../service/mocks --outpkg=mocks --with-expecter --filename=withdrawal_repository_mock.go

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByNumber(ctx context.Context, number string) (*model.Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses []model.OrderStatus) ([]*model.Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus, accrual float32) error
}

type BalanceRepository interface {
	Create(ctx context.Context, balance *model.Balance) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	Update(ctx context.Context, balance *model.Balance) error
	AddAccrual(ctx context.Context, userID uuid.UUID, amount float32) error
	Withdraw(ctx context.Context, userID uuid.UUID, amount float32) error
}

type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *model.Withdrawal) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Withdrawal, error)
	ExistsByOrderNumber(ctx context.Context, orderNumber string) (bool, error)
}
