package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
)

type OrderProcessor struct {
	orderRepo     repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualClient accrual.ClientInterface
	logger        *zap.Logger
}

func NewOrderProcessor(
	orderRepo repository.OrderRepository,
	balanceRepo repository.BalanceRepository,
	accrualClient accrual.ClientInterface,
	logger *zap.Logger,
) *OrderProcessor {
	return &OrderProcessor{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
		logger:        logger,
	}
}

func (p *OrderProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	p.logger.Info("Starting order processor")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Stopping order processor")
			return
		case <-ticker.C:
			p.processOrders(ctx)
		}
	}
}

func (p *OrderProcessor) processOrders(ctx context.Context) {
	p.logger.Debug("Processing orders")

	statuses := []model.OrderStatus{model.OrderStatusNew, model.OrderStatusProcessing}
	orders, err := p.orderRepo.GetOrdersByStatus(ctx, statuses)
	if err != nil {
		return
	}

	if len(orders) == 0 {
		p.logger.Debug("No orders to process")
		return
	}

	p.logger.Info("Found orders to process", zap.Int("count", len(orders)))

	for _, order := range orders {
		err := p.ProcessOrder(ctx, order.Number)
		if err != nil {
			p.logger.Error("Failed to process order",
				zap.String("orderNumber", order.Number),
				zap.Error(err))
		}
	}
}

func (p *OrderProcessor) ProcessOrder(ctx context.Context, orderNumber string) error {

	order, err := p.orderRepo.GetByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}

	if order.Status == model.OrderStatusProcessed || order.Status == model.OrderStatusInvalid {
		return nil
	}

	accrualResponse, err := p.accrualClient.GetOrderInfo(ctx, orderNumber)
	if err != nil {

		if rateLimitErr, ok := err.(*accrual.RateLimitError); ok && rateLimitErr.IsRetryable() {
			p.logger.Info("Rate limit hit for order",
				zap.String("orderNumber", orderNumber),
				zap.Duration("retryAfter", rateLimitErr.RetryAfter))
			return nil
		}

		p.logger.Error("Failed to get order info from accrual system",
			zap.String("orderNumber", orderNumber),
			zap.Error(err))
		return err
	}

	err = p.orderRepo.UpdateStatus(ctx, order.ID, accrualResponse.Status, accrualResponse.Accrual)
	if err != nil {
		p.logger.Error("Failed to update order status",
			zap.String("orderNumber", orderNumber),
			zap.Error(err))
		return err
	}

	if accrualResponse.Status == model.OrderStatusProcessed && accrualResponse.Accrual > 0 {
		err = p.balanceRepo.AddAccrual(ctx, order.UserID, accrualResponse.Accrual)
		if err != nil {
			p.logger.Error("Failed to add accrual to user balance",
				zap.String("orderNumber", orderNumber),
				zap.String("userID", order.UserID.String()),
				zap.Float32("accrual", accrualResponse.Accrual),
				zap.Error(err))
			return err
		}

		p.logger.Info("Added accrual to user balance",
			zap.String("orderNumber", orderNumber),
			zap.String("userID", order.UserID.String()),
			zap.Float32("accrual", accrualResponse.Accrual))
	}

	p.logger.Info("Processed order",
		zap.String("orderNumber", orderNumber),
		zap.String("status", string(accrualResponse.Status)),
		zap.Any("accrual", accrualResponse.Accrual))

	return nil
}
