package getbalanceusecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/pkg/getbalancepkg"
)

type Usecase struct {
	balanceRepo repository.BalanceRepository
	logger      *zap.Logger
}

func New(
	balanceRepo repository.BalanceRepository,
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

	balance, err := u.balanceRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		u.logger.Error("Failed to get balance for user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	response := getbalancepkg.Response{
		Current:   balance.CurrentBalance,
		Withdrawn: balance.WithdrawnBalance,
	}
	c.JSON(http.StatusOK, response)
}
