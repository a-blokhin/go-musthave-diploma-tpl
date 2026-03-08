package accrual

import (
	"context"
	"go-musthave-diploma-tpl/internal/model"
)

//go:generate mockery --name=ClientInterface --output=../service/mocks --outpkg=mocks --with-expecter --filename=accrual_client_mock.go

type ClientInterface interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*model.AccrualSystemResponse, error)
}
