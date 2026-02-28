package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/service/mocks"
)

func TestOrderProcessor_ProcessOrder(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name           string
		orderNumber    string
		orderStatus    model.OrderStatus
		mockSetup      func(*mocks.OrderRepository, *mocks.BalanceRepository, *mocks.ClientInterface)
		expectedError  bool
		expectedStatus model.OrderStatus
	}{
		{
			name:        "Successfully process order with accrual",
			orderNumber: "12345678903",
			orderStatus: model.OrderStatusNew,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderID := uuid.New()
				userID := uuid.New()
				accrual := float32(500.0)

				orderRepo.EXPECT().GetByNumber(ctx, "12345678903").Return(&model.Order{
					ID:     orderID,
					Number: "12345678903",
					UserID: userID,
					Status: model.OrderStatusNew,
				}, nil)

				accrualClient.EXPECT().GetOrderInfo(ctx, "12345678903").Return(&model.AccrualSystemResponse{
					Order:   "12345678903",
					Status:  model.OrderStatusProcessed,
					Accrual: accrual,
				}, nil)

				orderRepo.EXPECT().UpdateStatus(ctx, orderID, model.OrderStatusProcessed, accrual).Return(nil)
				balanceRepo.EXPECT().AddAccrual(ctx, userID, accrual).Return(nil)
			},
			expectedError:  false,
			expectedStatus: model.OrderStatusProcessed,
		},
		{
			name:        "Process order with no accrual",
			orderNumber: "12345678904",
			orderStatus: model.OrderStatusNew,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderID := uuid.New()
				userID := uuid.New()

				orderRepo.EXPECT().GetByNumber(ctx, "12345678904").Return(&model.Order{
					ID:     orderID,
					Number: "12345678904",
					UserID: userID,
					Status: model.OrderStatusNew,
				}, nil)

				accrualClient.EXPECT().GetOrderInfo(ctx, "12345678904").Return(&model.AccrualSystemResponse{
					Order:   "12345678904",
					Status:  model.OrderStatusProcessed,
					Accrual: 0,
				}, nil)

				orderRepo.EXPECT().UpdateStatus(ctx, orderID, model.OrderStatusProcessed, float32(0)).Return(nil)
			},
			expectedError:  false,
			expectedStatus: model.OrderStatusProcessed,
		},
		{
			name:        "Order already processed",
			orderNumber: "12345678905",
			orderStatus: model.OrderStatusProcessed,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderRepo.EXPECT().GetByNumber(ctx, "12345678905").Return(&model.Order{
					ID:     uuid.New(),
					Number: "12345678905",
					UserID: uuid.New(),
					Status: model.OrderStatusProcessed,
				}, nil)
			},
			expectedError:  false,
			expectedStatus: model.OrderStatusProcessed,
		},
		{
			name:        "Order is invalid",
			orderNumber: "12345678906",
			orderStatus: model.OrderStatusNew,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderID := uuid.New()
				userID := uuid.New()

				orderRepo.EXPECT().GetByNumber(ctx, "12345678906").Return(&model.Order{
					ID:     orderID,
					Number: "12345678906",
					UserID: userID,
					Status: model.OrderStatusNew,
				}, nil)

				accrualClient.EXPECT().GetOrderInfo(ctx, "12345678906").Return(&model.AccrualSystemResponse{
					Order:   "12345678906",
					Status:  model.OrderStatusInvalid,
					Accrual: 0,
				}, nil)

				orderRepo.EXPECT().UpdateStatus(ctx, orderID, model.OrderStatusInvalid, float32(0)).Return(nil)
			},
			expectedError:  false,
			expectedStatus: model.OrderStatusInvalid,
		},
		{
			name:        "Rate limit error",
			orderNumber: "12345678907",
			orderStatus: model.OrderStatusNew,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderRepo.EXPECT().GetByNumber(ctx, "12345678907").Return(&model.Order{
					ID:     uuid.New(),
					Number: "12345678907",
					UserID: uuid.New(),
					Status: model.OrderStatusNew,
				}, nil)

				retryAfter := 60 * time.Second
				accrualClient.EXPECT().GetOrderInfo(ctx, "12345678907").Return(nil, &accrual.RateLimitError{
					RetryAfter: retryAfter,
				})
			},
			expectedError:  false,
			expectedStatus: model.OrderStatusNew,
		},
		{
			name:        "Order not found",
			orderNumber: "12345678908",
			orderStatus: model.OrderStatusNew,
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderRepo.EXPECT().GetByNumber(ctx, "12345678908").Return(nil, errors.New("order not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderRepo := mocks.NewOrderRepository(t)
			mockBalanceRepo := mocks.NewBalanceRepository(t)
			mockAccrualClient := mocks.NewClientInterface(t)

			tt.mockSetup(mockOrderRepo, mockBalanceRepo, mockAccrualClient)

			processor := NewOrderProcessor(mockOrderRepo, mockBalanceRepo, mockAccrualClient, logger)

			err := processor.ProcessOrder(ctx, tt.orderNumber)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderProcessor_processOrders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		mockSetup func(*mocks.OrderRepository, *mocks.BalanceRepository, *mocks.ClientInterface)
	}{
		{
			name: "Process multiple orders",
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				userID := uuid.New()
				order1ID := uuid.New()
				order2ID := uuid.New()
				accrual := float32(500.0)

				orders := []*model.Order{
					{
						ID:     order1ID,
						Number: "12345678903",
						UserID: userID,
						Status: model.OrderStatusNew,
					},
					{
						ID:     order2ID,
						Number: "12345678904",
						UserID: userID,
						Status: model.OrderStatusProcessing,
					},
				}

				orderRepo.EXPECT().GetOrdersByStatus(ctx, []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}).Return(orders, nil)

				for _, order := range orders {
					orderRepo.EXPECT().GetByNumber(ctx, order.Number).Return(order, nil)
					accrualClient.EXPECT().GetOrderInfo(ctx, order.Number).Return(&model.AccrualSystemResponse{
						Order:   order.Number,
						Status:  model.OrderStatusProcessed,
						Accrual: accrual,
					}, nil)
					orderRepo.EXPECT().UpdateStatus(ctx, order.ID, model.OrderStatusProcessed, accrual).Return(nil)
					balanceRepo.EXPECT().AddAccrual(ctx, userID, accrual).Return(nil)
				}
			},
		},
		{
			name: "No orders to process",
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderRepo.EXPECT().GetOrdersByStatus(ctx, []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}).Return([]*model.Order{}, nil)
			},
		},
		{
			name: "Error getting orders",
			mockSetup: func(orderRepo *mocks.OrderRepository, balanceRepo *mocks.BalanceRepository, accrualClient *mocks.ClientInterface) {
				orderRepo.EXPECT().GetOrdersByStatus(ctx, []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}).Return(nil, errors.New("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockOrderRepo := mocks.NewOrderRepository(t)
			mockBalanceRepo := mocks.NewBalanceRepository(t)
			mockAccrualClient := mocks.NewClientInterface(t)

			tt.mockSetup(mockOrderRepo, mockBalanceRepo, mockAccrualClient)

			processor := NewOrderProcessor(mockOrderRepo, mockBalanceRepo, mockAccrualClient, logger)

			processor.processOrders(ctx)
		})
	}
}

func TestOrderProcessor_Start(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockOrderRepo := mocks.NewOrderRepository(t)
	mockBalanceRepo := mocks.NewBalanceRepository(t)
	mockAccrualClient := mocks.NewClientInterface(t)

	processor := NewOrderProcessor(mockOrderRepo, mockBalanceRepo, mockAccrualClient, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	processor.Start(ctx)

	assert.True(t, true)
}
