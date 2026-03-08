package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	passwordService := NewPasswordService(4)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Short password",
			password: "pass",
			wantErr:  true,
		},
		{
			name:     "Long password",
			password: "verylongpasswordwithspecialchars123!@#",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Password without digits",
			password: "password",
			wantErr:  true,
		},
		{
			name:     "Password without letters",
			password: "123456",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := passwordService.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if hashedPassword == "" {
					t.Error("HashPassword() returned empty hash")
				}
				if hashedPassword == tt.password {
					t.Error("HashPassword() returned unhashed password")
				}
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	passwordService := NewPasswordService(4)
	password := "testpassword123"
	hashedPassword, err := passwordService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "Correct password",
			hashedPassword: hashedPassword,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "Incorrect password",
			hashedPassword: hashedPassword,
			password:       "wrongpassword",
			wantErr:        true,
		},
		{
			name:           "Empty password",
			hashedPassword: hashedPassword,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "Invalid hash",
			hashedPassword: "invalidhash",
			password:       password,
			wantErr:        true,
		},
		{
			name:           "Empty hash",
			hashedPassword: "",
			password:       password,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := passwordService.CheckPassword(tt.password, tt.hashedPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPasswordHashingConsistency(t *testing.T) {
	passwordService := NewPasswordService(4)
	password := "consistencytest123"

	hash1, err1 := passwordService.HashPassword(password)
	hash2, err2 := passwordService.HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to hash password: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Error("Hashes of the same password should be different due to salt")
	}

	err1 = passwordService.CheckPassword(password, hash1)
	err2 = passwordService.CheckPassword(password, hash2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to check password: %v, %v", err1, err2)
	}
}
