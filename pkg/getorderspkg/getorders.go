package getorderspkg

import (
	"net/http"
	"time"
)

const MethodPath = "/api/user/orders"
const Method = http.MethodGet

type Response []OrderResponse

type OrderResponse struct {
	Number     string      `json:"number"`
	Status     string      `json:"status"`
	Accrual    float32     `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at"`
}