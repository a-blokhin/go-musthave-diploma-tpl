package model

import (
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserID           uuid.UUID `json:"user_id"`
	CurrentBalance   float32   `json:"current_balance"`
	WithdrawnBalance float32   `json:"withdrawn_balance"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewBalance(userID uuid.UUID) *Balance {
	return &Balance{
		UserID:           userID,
		CurrentBalance:   0,
		WithdrawnBalance: 0,
		UpdatedAt:        time.Now(),
	}
}

type BalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func (b *Balance) ToResponse() *BalanceResponse {
	return &BalanceResponse{
		Current:   b.CurrentBalance,
		Withdrawn: b.WithdrawnBalance,
	}
}

func (b *Balance) AddAccrual(amount float32) {
	b.CurrentBalance += amount
	b.UpdatedAt = time.Now()
}

func (b *Balance) Withdraw(amount float32) error {
	if b.CurrentBalance < amount {
		return ErrInsufficientBalance
	}
	b.CurrentBalance -= amount
	b.WithdrawnBalance += amount
	b.UpdatedAt = time.Now()
	return nil
}
