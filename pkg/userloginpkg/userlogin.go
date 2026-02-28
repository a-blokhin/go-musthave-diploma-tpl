package userloginpkg

import "net/http"

const MethodPath = "/api/user/login"
const Method = http.MethodPost

type Request struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Response struct {
	Message string `json:"message"`
}