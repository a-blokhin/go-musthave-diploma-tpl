package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewUser(login, passwordHash string) *User {
	return &User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
}

type UserRegistrationRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Login:     u.Login,
		CreatedAt: u.CreatedAt,
	}
}
