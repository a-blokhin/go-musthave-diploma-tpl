package getbalancepkg

import "net/http"

const MethodPath = "/api/user/balance"
const Method = http.MethodGet

type Response struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}