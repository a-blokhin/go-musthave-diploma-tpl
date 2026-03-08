package userloginusecase_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/usecase/userloginusecase"
	userloginmocks "go-musthave-diploma-tpl/internal/usecase/userloginusecase/mocks"
)

func TestUserLoginUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	passwordSvc := auth.NewPasswordService(4)
	hashedPassword, err := passwordSvc.HashPassword("password123")
	require.NoError(t, err)

	testUser := &model.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: hashedPassword,
	}

	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(*userloginmocks.UserRepository)
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "Successful login",
			requestBody: model.UserLoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {
				repo.EXPECT().GetByLogin(mock.Anything, "testuser").Return(testUser, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"message": "Login successful",
			},
		},
		{
			name: "User not found",
			requestBody: model.UserLoginRequest{
				Login:    "nonexistentuser",
				Password: "password123",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {
				repo.EXPECT().GetByLogin(mock.Anything, "nonexistentuser").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"error": "Invalid credentials",
			},
		},
		{
			name: "Database error on user lookup",
			requestBody: model.UserLoginRequest{
				Login:    "testuser",
				Password: "password123",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {
				repo.EXPECT().GetByLogin(mock.Anything, "testuser").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"error": "Invalid credentials",
			},
		},
		{
			name: "Invalid request format",
			requestBody: map[string]any{
				"login": "testuser",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
		{
			name: "Empty login",
			requestBody: model.UserLoginRequest{
				Login:    "",
				Password: "password123",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
		{
			name: "Empty password",
			requestBody: model.UserLoginRequest{
				Login:    "testuser",
				Password: "",
			},
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Invalid request format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := userloginmocks.NewUserRepository(t)

			tt.mockSetup(mockRepo)

			jwtSvc := auth.NewJWTService("test-secret")

			useCase := userloginusecase.New(mockRepo, passwordSvc, jwtSvc, logger)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			useCase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				if expectedMap, ok := tt.expectedBody.(map[string]any); ok {
					for key, expectedValue := range expectedMap {
						assert.Equal(t, expectedValue, response[key])
					}
				}
			}

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Bearer ", w.Header().Get("Authorization")[:7])
				assert.NotEmpty(t, w.Header().Get("Authorization")[7:])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserLoginUseCase_PasswordVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	passwordSvc := auth.NewPasswordService(4)
	jwtSvc := auth.NewJWTService("test-secret")

	password := "password123"
	hashedPassword, err := passwordSvc.HashPassword(password)
	require.NoError(t, err)

	testUser := &model.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: hashedPassword,
	}

	tests := []struct {
		name           string
		password       string
		expectedStatus int
	}{
		{
			name:           "Correct password",
			password:       password,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Incorrect password",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := userloginmocks.NewUserRepository(t)
			mockRepo.EXPECT().GetByLogin(mock.Anything, "testuser").Return(testUser, nil)

			useCase := userloginusecase.New(mockRepo, passwordSvc, jwtSvc, logger)

			body, _ := json.Marshal(model.UserLoginRequest{
				Login:    "testuser",
				Password: tt.password,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			useCase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Bearer ", w.Header().Get("Authorization")[:7])
				assert.NotEmpty(t, w.Header().Get("Authorization")[7:])
			} else {
				assert.Empty(t, w.Header().Get("Authorization"))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserLoginUseCase_TokenGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	passwordSvc := auth.NewPasswordService(4)
	jwtSvc := auth.NewJWTService("test-secret")

	password := "password123"
	hashedPassword, _ := passwordSvc.HashPassword(password)

	testUser := &model.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: hashedPassword,
	}

	mockRepo := userloginmocks.NewUserRepository(t)
	mockRepo.EXPECT().GetByLogin(mock.Anything, "testuser").Return(testUser, nil)

	useCase := userloginusecase.New(mockRepo, passwordSvc, jwtSvc, logger)

	body, _ := json.Marshal(model.UserLoginRequest{
		Login:    "testuser",
		Password: password,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	useCase.Execute(c)

	assert.Equal(t, http.StatusOK, w.Code)

	authHeader := w.Header().Get("Authorization")
	assert.Equal(t, "Bearer ", authHeader[:7])
	token := authHeader[7:]
	assert.NotEmpty(t, token)

	claims, err := jwtSvc.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID, claims.UserID)
	assert.Equal(t, testUser.Login, claims.Login)

	mockRepo.AssertExpectations(t)
}

func TestUserLoginUseCase_InputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	passwordSvc := auth.NewPasswordService(4)
	jwtSvc := auth.NewJWTService("test-secret")

	tests := []struct {
		name           string
		login          string
		password       string
		expectedStatus int
		mockSetup      func(*userloginmocks.UserRepository)
	}{
		{
			name:           "Valid input",
			login:          "validuser",
			password:       "validpassword",
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func(repo *userloginmocks.UserRepository) {
				repo.EXPECT().GetByLogin(mock.Anything, "validuser").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:           "Empty login",
			login:          "",
			password:       "validpassword",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
		},
		{
			name:           "Empty password",
			login:          "validuser",
			password:       "",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
		},
		{
			name:           "Whitespace login",
			login:          "   ",
			password:       "validpassword",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
		},
		{
			name:           "Whitespace password",
			login:          "validuser",
			password:       "   ",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(repo *userloginmocks.UserRepository) {

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := userloginmocks.NewUserRepository(t)

			tt.mockSetup(mockRepo)

			useCase := userloginusecase.New(mockRepo, passwordSvc, jwtSvc, logger)

			body, _ := json.Marshal(model.UserLoginRequest{
				Login:    tt.login,
				Password: tt.password,
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			useCase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockRepo.AssertExpectations(t)
		})
	}
}
