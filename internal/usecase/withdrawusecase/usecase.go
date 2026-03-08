package withdrawusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/pkg/withdrawpkg"
)

type Usecase struct {
	balanceRepo BalanceRepository
	logger      *zap.Logger
}

func New(
	balanceRepo BalanceRepository,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
		balanceRepo: balanceRepo,
		logger:      logger,
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

	var req withdrawpkg.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if !model.IsValidOrderNumber(req.Order) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		return
	}

	err = u.balanceRepo.WithdrawWithRecord(c.Request.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if err == model.ErrInsufficientBalance {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
			return
		}
		u.logger.Error("Failed to withdraw with record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, withdrawpkg.Response{Message: "Withdrawal successful"})
}
