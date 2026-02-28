package getwithdrawalspkg

import (
	"net/http"
	"time"
)

const MethodPath = "/api/user/withdrawals"
const Method = http.MethodGet

type Response []WithdrawalResponse

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}