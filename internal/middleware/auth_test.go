package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go-musthave-diploma-tpl/internal/auth"
)

func TestJWTAuthMiddleware(t *testing.T) {
	secretKey := "test-secret-key"
	logger := zaptest.NewLogger(t)

	jwtService := auth.NewJWTService(secretKey)

	userID := uuid.New()
	login := "testuser"
	validToken, err := jwtService.GenerateToken(userID, login)
	require.NoError(t, err)

	expiredToken := generateExpiredTestToken(userID, login, secretKey)

	invalidToken := "invalid.token.format"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
		expectedLogin  string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: userID.String(),
			expectedLogin:  login,
		},
		{
			name:           "No authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
			expectedLogin:  "",
		},
		{
			name:           "Invalid authorization header format",
			authHeader:     "InvalidFormat " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
			expectedLogin:  "",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer " + invalidToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
			expectedLogin:  "",
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
			expectedLogin:  "",
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
			expectedLogin:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(JWTAuthMiddleware(jwtService, logger))

			router.GET("/test", func(c *gin.Context) {
				if tt.expectedUserID != "" {
					userID, exists := GetUserID(c)
					require.True(t, exists)
					assert.Equal(t, tt.expectedUserID, userID)

					userLogin, exists := GetUserLogin(c)
					require.True(t, exists)
					assert.Equal(t, tt.expectedLogin, userLogin)
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	secretKey := "test-secret-key"
	logger := zaptest.NewLogger(t)

	jwtService := auth.NewJWTService(secretKey)

	userID := uuid.New()
	login := "testuser"
	validToken, err := jwtService.GenerateToken(userID, login)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
		expectedLogin  string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: userID.String(),
			expectedLogin:  login,
		},
		{
			name:           "No authorization header",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedUserID: "",
			expectedLogin:  "",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.format",
			expectedStatus: http.StatusOK,
			expectedUserID: "",
			expectedLogin:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(OptionalAuthMiddleware(jwtService, logger))

			router.GET("/test", func(c *gin.Context) {
				if tt.expectedUserID != "" {
					userID, exists := GetUserID(c)
					require.True(t, exists)
					assert.Equal(t, tt.expectedUserID, userID)

					userLogin, exists := GetUserLogin(c)
					require.True(t, exists)
					assert.Equal(t, tt.expectedLogin, userLogin)
				} else {

					_, exists := GetUserID(c)
					assert.False(t, exists)

					_, exists = GetUserLogin(c)
					assert.False(t, exists)
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetUserIDUUID(t *testing.T) {

	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("userID", userID.String())

	retrievedUserID, exists := GetUserIDUUID(c)
	require.True(t, exists)
	assert.Equal(t, userID, retrievedUserID)
}

func TestGetUserIDUUID_InvalidUUID(t *testing.T) {

	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("userID", "invalid-uuid")

	retrievedUserID, exists := GetUserIDUUID(c)
	require.False(t, exists)
	assert.Equal(t, uuid.Nil, retrievedUserID)
}

func TestGetUserIDUUID_NotSet(t *testing.T) {

	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	retrievedUserID, exists := GetUserIDUUID(c)
	require.False(t, exists)
	assert.Equal(t, uuid.Nil, retrievedUserID)
}

func generateExpiredTestToken(userID uuid.UUID, login, secretKey string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"login":   login,
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secretKey))
	return tokenString
}
