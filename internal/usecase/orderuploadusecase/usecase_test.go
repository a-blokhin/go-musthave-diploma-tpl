package orderuploadusecase_test

import (
	"bytes"
	"context"
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

	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/usecase/orderuploadusecase"
	orderuploadmocks "go-musthave-diploma-tpl/internal/usecase/orderuploadusecase/mocks"
)

func TestOrderUploadUseCase_Execute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*orderuploadmocks.OrderRepository)
		expectedStatus int
		expectedBody   any
	}{
		{
			name:        "Successful new order upload",
			requestBody: "12345678903",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(nil, model.ErrOrderNotFound)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody: map[string]any{
				"message": "Order accepted for processing",
			},
		},
		{
			name:        "Order already uploaded by different user",
			requestBody: "12345678903",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				otherUserID := uuid.New()
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(&model.Order{UserID: otherUserID}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]any{
				"error": "Order already uploaded by another user",
			},
		},
		{
			name:        "Order already uploaded by same user",
			requestBody: "12345678903",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(&model.Order{UserID: userID}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"message": "Order already uploaded",
			},
		},
		{
			name:        "Database error on checking order exists",
			requestBody: "12345678903",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"error": http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name:        "Database error on order creation",
			requestBody: "12345678903",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(nil, model.ErrOrderNotFound)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Order")).Return(errors.New("creation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"error": http.StatusText(http.StatusInternalServerError),
			},
		},
		{
			name:        "Invalid order number",
			requestBody: "12345678904",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: map[string]any{
				"error": "Invalid order number format",
			},
		},
		{
			name:        "Empty order number",
			requestBody: "",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Order number is required",
			},
		},
		{
			name:        "Whitespace only order number",
			requestBody: "   ",
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Order number is required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := orderuploadmocks.NewOrderRepository(t)

			tt.mockSetup(mockRepo)

			useCase := orderuploadusecase.New(mockRepo, logger)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/orders", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "text/plain")
			c.Set("userID", userID.String())

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

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderUploadUseCase_OrderValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		orderNumber    string
		expectedStatus int
		mockSetup      func(*orderuploadmocks.OrderRepository)
	}{
		{
			name:           "Valid order number",
			orderNumber:    "12345678903",
			expectedStatus: http.StatusAccepted,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "12345678903").Return(nil, model.ErrOrderNotFound)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
			},
		},
		{
			name:           "Invalid order number - wrong checksum",
			orderNumber:    "12345678904",
			expectedStatus: http.StatusUnprocessableEntity,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
		},
		{
			name:           "Empty order number",
			orderNumber:    "",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
		},
		{
			name:           "Non-numeric order number",
			orderNumber:    "abc123",
			expectedStatus: http.StatusUnprocessableEntity,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {

			},
		},
		{
			name:           "Another valid order number",
			orderNumber:    "9278923470",
			expectedStatus: http.StatusAccepted,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "9278923470").Return(nil, model.ErrOrderNotFound)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
			},
		},
		{
			name:           "Short valid order number",
			orderNumber:    "18",
			expectedStatus: http.StatusAccepted,
			mockSetup: func(repo *orderuploadmocks.OrderRepository) {
				repo.EXPECT().GetByNumber(mock.Anything, "18").Return(nil, model.ErrOrderNotFound)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := orderuploadmocks.NewOrderRepository(t)

			tt.mockSetup(mockRepo)

			useCase := orderuploadusecase.New(mockRepo, logger)
			userID := uuid.New()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/orders", bytes.NewBufferString(tt.orderNumber))
			c.Request.Header.Set("Content-Type", "text/plain")
			c.Set("userID", userID.String())

			useCase.Execute(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestOrderUploadUseCase_OrderCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockRepo := orderuploadmocks.NewOrderRepository(t)
	useCase := orderuploadusecase.New(mockRepo, logger)

	userID := uuid.New()
	orderNumber := "12345678903"

	mockRepo.EXPECT().GetByNumber(mock.Anything, orderNumber).Return(nil, model.ErrOrderNotFound)
	mockRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(order *model.Order) bool {
		return order.Number == orderNumber && order.UserID == userID && order.Status == model.OrderStatusNew
	})).Return(nil).Run(func(ctx context.Context, order *model.Order) {
		assert.Equal(t, orderNumber, order.Number)
		assert.Equal(t, userID, order.UserID)
		assert.Equal(t, model.OrderStatusNew, order.Status)
		assert.Zero(t, order.Accrual)
		assert.NotEmpty(t, order.UploadedAt)
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/orders", bytes.NewBufferString(orderNumber))
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Set("userID", userID.String())

	useCase.Execute(c)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Order accepted for processing", response["message"])

	mockRepo.AssertExpectations(t)
}

func TestOrderUploadUseCase_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockRepo := orderuploadmocks.NewOrderRepository(t)
	useCase := orderuploadusecase.New(mockRepo, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/orders", bytes.NewBufferString("12345678903"))
	c.Request.Header.Set("Content-Type", "text/plain")

	useCase.Execute(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "User not authenticated", response["error"])
}

func TestOrderUploadUseCase_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zaptest.NewLogger(t)

	mockRepo := orderuploadmocks.NewOrderRepository(t)
	useCase := orderuploadusecase.New(mockRepo, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/orders", bytes.NewBufferString("12345678903"))
	c.Request.Header.Set("Content-Type", "text/plain")
	c.Set("userID", "invalid-uuid")

	useCase.Execute(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid user ID", response["error"])
}
