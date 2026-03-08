package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/model"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*model.AccrualSystemResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseSuccessResponse(resp)
	case http.StatusNoContent:
		return nil, fmt.Errorf("order not registered in accrual system")
	case http.StatusTooManyRequests:
		retryAfter := c.getRetryAfter(resp)
		return nil, &RateLimitError{
			RetryAfter: retryAfter,
			Message:    "Rate limit exceeded",
		}
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("accrual system internal error")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) parseSuccessResponse(resp *http.Response) (*model.AccrualSystemResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var accrualResponse model.AccrualSystemResponse
	if err := json.Unmarshal(body, &accrualResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	switch model.AccrualSystemStatus(accrualResponse.Status) {
	case model.AccrualStatusRegistered:
		accrualResponse.Status = model.OrderStatusNew
	case model.AccrualStatusProcessing:
		accrualResponse.Status = model.OrderStatusProcessing
	case model.AccrualStatusInvalid:
		accrualResponse.Status = model.OrderStatusInvalid
	case model.AccrualStatusProcessed:
		accrualResponse.Status = model.OrderStatusProcessed
	default:
		accrualResponse.Status = model.OrderStatusNew
	}

	return &accrualResponse, nil
}

func (c *Client) getRetryAfter(resp *http.Response) time.Duration {
	retryAfterStr := resp.Header.Get("Retry-After")
	if retryAfterStr == "" {
		return 60 * time.Second
	}

	retryAfter, err := strconv.Atoi(retryAfterStr)
	if err != nil {
		c.logger.Warn("Invalid Retry-After header", zap.String("value", retryAfterStr), zap.Error(err))
		return 60 * time.Second
	}

	return time.Duration(retryAfter) * time.Second
}

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s: retry after %v", e.Message, e.RetryAfter)
}

func (e *RateLimitError) IsRetryable() bool {
	return true
}
