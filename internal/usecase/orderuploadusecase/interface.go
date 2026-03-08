package orderuploadusecase

import (
	"context"

	"go-musthave-diploma-tpl/internal/model"
)

//go:generate mockery --name=OrderRepository --output=./mocks --outpkg=mocks --with-expecter --filename=order_repository_mock.go

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByNumber(ctx context.Context, number string) (*model.Order, error)
}
