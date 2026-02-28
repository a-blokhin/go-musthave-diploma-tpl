package model

import (
	"time"

	"github.com/google/uuid"
)

type Withdrawal struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	UserID      uuid.UUID `json:"user_id"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func NewWithdrawal(orderNumber string, userID uuid.UUID, sum float32) *Withdrawal {
	return &Withdrawal{
		ID:          uuid.New(),
		OrderNumber: orderNumber,
		UserID:      userID,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}
}

type WithdrawalRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float32 `json:"sum" binding:"required,gt=0"`
}

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w *Withdrawal) ToResponse() *WithdrawalResponse {
	return &WithdrawalResponse{
		Order:       w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}
}
