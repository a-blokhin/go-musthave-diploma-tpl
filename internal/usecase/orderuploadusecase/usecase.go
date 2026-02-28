package orderuploadusecase

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/pkg/orderuploadpkg"
)

type Usecase struct {
	orderRepo OrderRepository
	logger    *zap.Logger
}

func New(
	orderRepo OrderRepository,
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

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	defer c.Request.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order number is required"})
		return
	}

	if !model.IsValidOrderNumber(orderNumber) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid order number format"})
		return
	}

	existingOrder, err := u.orderRepo.GetByNumber(c.Request.Context(), orderNumber)
	if err != nil {
		if !errors.Is(err, model.ErrOrderNotFound) {
			u.logger.Error("Failed to check if order exists", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
	} else if existingOrder != nil {
		if existingOrder.UserID == userID {
			c.JSON(http.StatusOK, orderuploadpkg.Response{
				Message: "Order already uploaded",
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Order already uploaded by another user"})
		return
	}

	order := model.NewOrder(orderNumber, userID)
	err = u.orderRepo.Create(c.Request.Context(), order)
	if err != nil {
		if errors.Is(err, model.ErrOrderAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Order already uploaded by another user"})
			return
		}

		u.logger.Error("Failed to create order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusAccepted, orderuploadpkg.Response{
		Message: "Order accepted for processing",
	})
}
