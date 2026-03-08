package userregistrationusecase

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/pkg/userregisterpkg"
)

type Usecase struct {
	userRepo    UserRepository
	passwordSvc *auth.PasswordService
	jwtSvc      *auth.JWTService
	balanceRepo BalanceRepository
	logger      *zap.Logger
}

func New(
	userRepo UserRepository,
	passwordSvc *auth.PasswordService,
	jwtSvc *auth.JWTService,
	balanceRepo BalanceRepository,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
		userRepo:    userRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
		balanceRepo: balanceRepo,
		logger:      logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	var req userregisterpkg.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	hashedPassword, err := u.passwordSvc.HashPassword(req.Password)
	if err != nil {
		u.logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	user := model.NewUser(req.Login, hashedPassword)
	err = u.userRepo.Create(c.Request.Context(), user)
	if err != nil {
		if errors.Is(err, model.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Login already exists"})
			return
		}

		u.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	balance := model.NewBalance(user.ID)
	err = u.balanceRepo.Create(c.Request.Context(), balance)
	if err != nil {
		u.logger.Error("Failed to create user balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	token, err := u.jwtSvc.GenerateToken(user.ID, user.Login)
	if err != nil {
		u.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.Header("Authorization", "Bearer "+token)

	response := userregisterpkg.Response{
		Login: user.Login,
	}
	c.JSON(http.StatusOK, response)
}
