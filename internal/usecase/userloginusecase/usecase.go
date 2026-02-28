package userloginusecase

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/pkg/userloginpkg"
)

type Usecase struct {
	userRepo    UserRepository
	passwordSvc *auth.PasswordService
	jwtSvc      *auth.JWTService
	logger      *zap.Logger
}

func New(
	userRepo UserRepository,
	passwordSvc *auth.PasswordService,
	jwtSvc *auth.JWTService,
	logger *zap.Logger,
) *Usecase {
	return &Usecase{
		userRepo:    userRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
		logger:      logger,
	}
}

func (u *Usecase) Execute(c *gin.Context) {
	var req userloginpkg.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(strings.TrimSpace(req.Login)) == 0 || len(strings.TrimSpace(req.Password)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := u.userRepo.GetByLogin(c.Request.Context(), req.Login)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = u.passwordSvc.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := u.jwtSvc.GenerateToken(user.ID, user.Login)
	if err != nil {
		u.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.Header("Authorization", "Bearer "+token)

	c.JSON(http.StatusOK, userloginpkg.Response{
		Message: "Login successful",
	})
}
