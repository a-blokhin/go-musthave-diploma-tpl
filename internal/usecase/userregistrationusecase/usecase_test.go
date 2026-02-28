package userregistrationusecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/usecase/userregistrationusecase"
	userregistrationmocks "go-musthave-diploma-tpl/internal/usecase/userregistrationusecase/mocks"
)

func TestUserRegistrationUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(*userregistrationmocks.UserRepository, *userregistrationmocks.BalanceRepository)
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "Successful registration",
			requestBody: model.UserRegistrationRequest{
				Login:    "testuser",
				Password: "password123",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
				balanceRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Balance")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"login": "testuser",
			},
		},
		{
			name: "User already exists",
			requestBody: model.UserRegistrationRequest{
				Login:    "existinguser",
				Password: "password123",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).Return(model.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]any{
				"error": "Login already exists",
			},
		},
		{
			name: "Database error on user creation",
			requestBody: model.UserRegistrationRequest{
				Login:    "testuser",
				Password: "password123",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).Return(errors.New("creation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"error": http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name: "Invalid request format",
			requestBody: map[string]any{
				"login": "testuser",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
		{
			name: "Empty login",
			requestBody: model.UserRegistrationRequest{
				Login:    "",
				Password: "password123",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
		{
			name: "Empty password",
			requestBody: model.UserRegistrationRequest{
				Login:    "testuser",
				Password: "",
			},
			mockSetup: func(userRepo *userregistrationmocks.UserRepository, balanceRepo *userregistrationmocks.BalanceRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockUserRepo := userregistrationmocks.NewUserRepository(t)
			mockBalanceRepo := userregistrationmocks.NewBalanceRepository(t)

			tt.mockSetup(mockUserRepo, mockBalanceRepo)

			passwordSvc := auth.NewPasswordService(4)
			jwtSvc := auth.NewJWTService("test-secret")

			useCase := userregistrationusecase.New(mockUserRepo, passwordSvc, jwtSvc, mockBalanceRepo, logger)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			useCase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if expectedMap, ok := tt.expectedBody.(map[string]any); ok {
					for key, expectedValue := range expectedMap {
						if key == "login" {
							assert.Equal(t, expectedValue, response[key])
						} else if key == "error" {
							assert.Equal(t, expectedValue, response[key])
						}
					}
				}
			}

			mockUserRepo.AssertExpectations(t)
			mockBalanceRepo.AssertExpectations(t)
		})
	}
}

func TestUserRegistrationUseCase_PasswordHashing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockUserRepo := userregistrationmocks.NewUserRepository(t)
	mockBalanceRepo := userregistrationmocks.NewBalanceRepository(t)

	mockUserRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Run(func(ctx context.Context, user *model.User) {
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEqual(t, "password123", user.PasswordHash)
	})
	mockBalanceRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Balance")).Return(nil)

	passwordSvc := auth.NewPasswordService(4)
	jwtSvc := auth.NewJWTService("test-secret")

	useCase := userregistrationusecase.New(mockUserRepo, passwordSvc, jwtSvc, mockBalanceRepo, logger)

	body, _ := json.Marshal(model.UserRegistrationRequest{
		Login:    "testuser",
		Password: "password123",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	useCase.Execute(c)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, "Bearer ", w.Header().Get("Authorization")[:7])
	assert.NotEmpty(t, w.Header().Get("Authorization")[7:])

	mockUserRepo.AssertExpectations(t)
	mockBalanceRepo.AssertExpectations(t)
}
