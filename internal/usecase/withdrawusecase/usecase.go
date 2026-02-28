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
	balanceRepo    BalanceRepository
	withdrawalRepo WithdrawalRepository
	logger         *zap.Logger
}

func New(
	balanceRepo BalanceRepository,
	withdrawalRepo WithdrawalRepository,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
		logger:         logger,
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

	withdrawalExists, err := u.withdrawalRepo.ExistsByOrderNumber(c.Request.Context(), req.Order)
	if err != nil {
		u.logger.Error("Failed to check if withdrawal exists", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if withdrawalExists {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Withdrawal with this order number already exists"})
		return
	}

	balance, err := u.balanceRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get balance for user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if balance.CurrentBalance < req.Sum {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		return
	}

	withdrawal := model.NewWithdrawal(req.Order, userID, req.Sum)
	err = u.withdrawalRepo.Create(c.Request.Context(), withdrawal)
	if err != nil {
		u.logger.Error("Failed to create withdrawal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	err = u.balanceRepo.Withdraw(c.Request.Context(), userID, req.Sum)
	if err != nil {
		u.logger.Error("Failed to withdraw from balance", zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, withdrawpkg.Response{Message: "Withdrawal successful"})
}
