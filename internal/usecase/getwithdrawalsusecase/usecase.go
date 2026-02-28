package getwithdrawalsusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/pkg/getwithdrawalspkg"
)

type Usecase struct {
	withdrawalRepo repository.WithdrawalRepository
	logger         *zap.Logger
}

func New(
	withdrawalRepo repository.WithdrawalRepository,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
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

	withdrawals, err := u.withdrawalRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get withdrawals for user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	responses := make(getwithdrawalspkg.Response, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		responses = append(responses, getwithdrawalspkg.WithdrawalResponse{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	if len(responses) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, responses)
}
