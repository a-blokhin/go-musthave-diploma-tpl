package auth

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	minLength int
}

func NewPasswordService(minLength int) *PasswordService {
	return &PasswordService{
		minLength: minLength,
	}
}

func (p *PasswordService) HashPassword(password string) (string, error) {
	if err := p.ValidatePassword(password); err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

func (p *PasswordService) CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (p *PasswordService) ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < p.minLength {
		return fmt.Errorf("password must be at least %d characters long", p.minLength)
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func DefaultPasswordService() *PasswordService {
	return NewPasswordService(8)
}
