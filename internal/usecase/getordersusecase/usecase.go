package getordersusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/pkg/getorderspkg"
)

type Usecase struct {
	orderRepo repository.OrderRepository
	logger    *zap.Logger
}

func New(
	orderRepo repository.OrderRepository,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	orders, err := u.orderRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get orders for user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	responses := make(getorderspkg.Response, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, getorderspkg.OrderResponse{
			Number:     order.Number,
			Status:     string(order.Status),
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	if len(responses) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, responses)
}
