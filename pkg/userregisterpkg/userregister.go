package userregisterpkg

import "net/http"

const MethodPath = "/api/user/register"
const Method = http.MethodPost

type Request struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Response struct {
	Login string `json:"login"`
}