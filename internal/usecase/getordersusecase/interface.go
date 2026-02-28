package getordersusecase

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"

	"github.com/google/uuid"
)

//go:generate mockery --name=OrderRepository --output=./mocks --outpkg=mocks --with-expecter --filename=order_repository_mock.go

type OrderRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
}
