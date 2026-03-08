package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/auth"
)

const (
	userIDContextKey    = "userID"
	userLoginContextKey = "userLogin"
	authHeaderKey       = "Authorization"
	bearerPrefix        = "Bearer "
)

func JWTAuthMiddleware(jwtService *auth.JWTService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid JWT token",
				zap.Error(err))

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		c.Set(userIDContextKey, claims.UserID.String())
		c.Set(userLoginContextKey, claims.Login)

		logger.Debug("Validated JWT token",
			zap.String("userID", claims.UserID.String()),
			zap.String("login", claims.Login))

		c.Next()
	}
}

func OptionalAuthMiddleware(jwtService *auth.JWTService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authHeaderKey)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.Warn("Invalid authorization header format in optional auth")
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			logger.Warn("Invalid JWT token in optional auth",
				zap.Error(err))
			c.Next()
			return
		}

		c.Set(userIDContextKey, claims.UserID.String())
		c.Set(userLoginContextKey, claims.Login)

		logger.Debug("Validated JWT token in optional auth",
			zap.String("userID", claims.UserID.String()),
			zap.String("login", claims.Login))

		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDContextKey)
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

func GetUserLogin(c *gin.Context) (string, bool) {
	userLogin, exists := c.Get(userLoginContextKey)
	if !exists {
		return "", false
	}

	login, ok := userLogin.(string)
	return login, ok
}

func GetUserIDUUID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, exists := GetUserID(c)
	if !exists {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}
