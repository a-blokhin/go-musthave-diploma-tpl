package withdrawpkg

import "net/http"

const MethodPath = "/api/user/balance/withdraw"
const Method = http.MethodPost

type Request struct {
	Order string  `json:"order" binding:"required"`
	Sum   float32 `json:"sum" binding:"required,gt=0"`
}

type Response struct {
	Message string `json:"message"`
}