package orderuploadpkg

import "net/http"

const MethodPath = "/api/user/orders"
const Method = http.MethodPost

type Request string

type Response struct {
	Message string `json:"message"`
}