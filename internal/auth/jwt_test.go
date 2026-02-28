package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		login     string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "Valid token generation",
			userID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			login:     "testuser",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "Empty secret key",
			userID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			login:     "testuser",
			secretKey: "",
			wantErr:   false,
		},
		{
			name:      "Empty login",
			userID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			login:     "",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := NewJWTService(tt.secretKey)
			token, err := jwtService.GenerateToken(tt.userID, tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	validUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	validLogin := "testuser"

	jwtService := NewJWTService(secretKey)
	validToken, err := jwtService.GenerateToken(validUserID, validLogin)
	if err != nil {
		t.Fatalf("Failed to generate valid token for test: %v", err)
	}

	expiredToken, err := generateExpiredToken(validUserID, validLogin, secretKey)
	if err != nil {
		t.Fatalf("Failed to generate expired token for test: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secretKey string
		wantErr   bool
		wantID    uuid.UUID
		wantLogin string
	}{
		{
			name:      "Valid token",
			token:     validToken,
			secretKey: secretKey,
			wantErr:   false,
			wantID:    validUserID,
			wantLogin: validLogin,
		},
		{
			name:      "Invalid token format",
			token:     "invalid.token.format",
			secretKey: secretKey,
			wantErr:   true,
		},
		{
			name:      "Empty token",
			token:     "",
			secretKey: secretKey,
			wantErr:   true,
		},
		{
			name:      "Wrong secret key",
			token:     validToken,
			secretKey: "wrong-secret-key",
			wantErr:   true,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			secretKey: secretKey,
			wantErr:   true,
		},
		{
			name:      "Token with different secret",
			token:     generateTokenWithDifferentSecret(validUserID, validLogin),
			secretKey: secretKey,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJwtService := NewJWTService(tt.secretKey)
			claims, err := testJwtService.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.UserID != tt.wantID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, tt.wantID)
				}
				if claims.Login != tt.wantLogin {
					t.Errorf("ValidateToken() login = %v, want %v", claims.Login, tt.wantLogin)
				}
			}
		})
	}
}

func TestTokenRoundTrip(t *testing.T) {
	secretKey := "roundtrip-secret-key"
	testCases := []struct {
		userID uuid.UUID
		login  string
	}{
		{uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), "user1"},
		{uuid.MustParse("123e4567-e89b-12d3-a456-426614174999"), "user999"},
		{uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), "zero-user"},
	}

	for _, tc := range testCases {
		t.Run("RoundTrip_"+tc.login, func(t *testing.T) {
			jwtService := NewJWTService(secretKey)

			token, err := jwtService.GenerateToken(tc.userID, tc.login)
			if err != nil {
				t.Fatalf("GenerateToken() error: %v", err)
			}

			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() error: %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, tc.userID)
			}
			if claims.Login != tc.login {
				t.Errorf("Login mismatch: got %s, want %s", claims.Login, tc.login)
			}
		})
	}
}

func generateExpiredToken(userID uuid.UUID, login, secretKey string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"login":   login,
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}

func generateTokenWithDifferentSecret(userID uuid.UUID, login string) string {
	differentSecret := "different-secret-key"
	jwtService := NewJWTService(differentSecret)
	token, _ := jwtService.GenerateToken(userID, login)
	return token
}
