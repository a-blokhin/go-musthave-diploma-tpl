package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         uuid.UUID   `json:"id"`
	Number     string      `json:"number"`
	UserID     uuid.UUID   `json:"user_id"`
	Status     OrderStatus `json:"status"`
	Accrual    float32     `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

func NewOrder(number string, userID uuid.UUID) *Order {
	return &Order{
		ID:         uuid.New(),
		Number:     number,
		UserID:     userID,
		Status:     OrderStatusNew,
		UploadedAt: time.Now(),
	}
}

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    float32     `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

func (o *Order) ToResponse() *OrderResponse {
	return &OrderResponse{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt,
	}
}

type AccrualSystemResponse struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float32     `json:"accrual"`
}

type AccrualSystemStatus string

const (
	AccrualStatusRegistered AccrualSystemStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualSystemStatus = "INVALID"
	AccrualStatusProcessing AccrualSystemStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualSystemStatus = "PROCESSED"
)

func (s AccrualSystemStatus) ToOrderStatus() OrderStatus {
	switch s {
	case AccrualStatusRegistered:
		return OrderStatusNew
	case AccrualStatusProcessing:
		return OrderStatusProcessing
	case AccrualStatusInvalid:
		return OrderStatusInvalid
	case AccrualStatusProcessed:
		return OrderStatusProcessed
	default:
		return OrderStatusNew
	}
}
