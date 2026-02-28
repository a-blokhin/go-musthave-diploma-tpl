package accrual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"go-musthave-diploma-tpl/internal/model"
)

func TestClient_GetOrderInfo(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name           string
		orderNumber    string
		responseStatus int
		responseBody   any
		expectedResult *model.AccrualSystemResponse
		expectedError  bool
	}{
		{
			name:           "Successful processed order",
			orderNumber:    "12345678903",
			responseStatus: http.StatusOK,
			responseBody: model.AccrualSystemResponse{
				Order:   "12345678903",
				Status:  model.OrderStatusProcessed,
				Accrual: 500.0,
			},
			expectedResult: &model.AccrualSystemResponse{
				Order:   "12345678903",
				Status:  model.OrderStatusProcessed,
				Accrual: 500.0,
			},
			expectedError: false,
		},
		{
			name:           "Order with processing status",
			orderNumber:    "12345678904",
			responseStatus: http.StatusOK,
			responseBody: model.AccrualSystemResponse{
				Order:  "12345678904",
				Status: model.OrderStatusProcessing,
			},
			expectedResult: &model.AccrualSystemResponse{
				Order:  "12345678904",
				Status: model.OrderStatusProcessing,
			},
			expectedError: false,
		},
		{
			name:           "Order not found",
			orderNumber:    "99999999999",
			responseStatus: http.StatusNoContent,
			responseBody:   nil,
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:           "Invalid order status",
			orderNumber:    "12345678905",
			responseStatus: http.StatusOK,
			responseBody: model.AccrualSystemResponse{
				Order:  "12345678905",
				Status: model.OrderStatusInvalid,
			},
			expectedResult: &model.AccrualSystemResponse{
				Order:  "12345678905",
				Status: model.OrderStatusInvalid,
			},
			expectedError: false,
		},
		{
			name:           "Rate limited",
			orderNumber:    "12345678906",
			responseStatus: http.StatusTooManyRequests,
			responseBody:   "No more than N requests per minute allowed",
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:           "Server error",
			orderNumber:    "12345678907",
			responseStatus: http.StatusInternalServerError,
			responseBody:   http.StatusText(http.StatusInternalServerError),
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/"+tt.orderNumber, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, logger)

			result, err := client.GetOrderInfo(context.Background(), tt.orderNumber)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_GetOrderInfoWithContext(t *testing.T) {
	logger := zaptest.NewLogger(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.AccrualSystemResponse{
			Order:   "12345678903",
			Status:  model.OrderStatusProcessed,
			Accrual: 500.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	result, err := client.GetOrderInfo(ctx, "12345678903")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_RateLimitHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount > 2 {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("No more than N requests per minute allowed"))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.AccrualSystemResponse{
			Order:   "12345678903",
			Status:  model.OrderStatusProcessed,
			Accrual: 500.0,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)

	for i := 0; i < 2; i++ {
		result, err := client.GetOrderInfo(context.Background(), "12345678903")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	}

	result, err := client.GetOrderInfo(context.Background(), "12345678903")
	assert.Error(t, err)
	assert.IsType(t, &RateLimitError{}, err)
	assert.Nil(t, result)
}

func TestClient_InvalidResponse(t *testing.T) {
	logger := zaptest.NewLogger(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	client := NewClient(server.URL, logger)

	result, err := client.GetOrderInfo(context.Background(), "12345678903")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestClient_NetworkError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	client := NewClient("http://invalid-url-that-does-not-exist.com", logger)

	result, err := client.GetOrderInfo(context.Background(), "12345678903")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8000"
	logger := zaptest.NewLogger(t)

	client := NewClient(baseURL, logger)

	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, logger, client.logger)
}

func TestRateLimitError(t *testing.T) {
	retryAfter := 60 * time.Second
	message := "Rate limit exceeded"

	err := &RateLimitError{
		RetryAfter: retryAfter,
		Message:    message,
	}

	assert.Equal(t, message+": retry after 1m0s", err.Error())
	assert.True(t, err.IsRetryable())
}

func TestAccrualSystemStatus_ToOrderStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   model.AccrualSystemStatus
		expected model.OrderStatus
	}{
		{
			name:     "REGISTERED to NEW",
			status:   model.AccrualStatusRegistered,
			expected: model.OrderStatusNew,
		},
		{
			name:     "PROCESSING to PROCESSING",
			status:   model.AccrualStatusProcessing,
			expected: model.OrderStatusProcessing,
		},
		{
			name:     "INVALID to INVALID",
			status:   model.AccrualStatusInvalid,
			expected: model.OrderStatusInvalid,
		},
		{
			name:     "PROCESSED to PROCESSED",
			status:   model.AccrualStatusProcessed,
			expected: model.OrderStatusProcessed,
		},
		{
			name:     "Unknown status to NEW",
			status:   model.AccrualSystemStatus("UNKNOWN"),
			expected: model.OrderStatusNew,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.ToOrderStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

