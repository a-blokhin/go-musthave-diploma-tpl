package model

import (
	"testing"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid card number",
			number:   "4532015112830366",
			expected: true,
		},
		{
			name:     "Valid order number from requirements",
			number:   "12345678903",
			expected: true,
		},
		{
			name:     "Another valid order number",
			number:   "9278923470",
			expected: true,
		},
		{
			name:     "Invalid card number - wrong checksum",
			number:   "4532015112830367",
			expected: false,
		},
		{
			name:     "Invalid order number",
			number:   "12345678904",
			expected: false,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: false,
		},
		{
			name:     "Non-numeric string",
			number:   "abc123",
			expected: false,
		},
		{
			name:     "Single digit",
			number:   "0",
			expected: true,
		},
		{
			name:     "Short valid number",
			number:   "18",
			expected: true,
		},
		{
			name:     "Long valid number",
			number:   "79927398713",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLuhn(tt.number)
			if result != tt.expected {
				t.Errorf("ValidateLuhn(%s) = %v; want %v", tt.number, result, tt.expected)
			}
		})
	}
}

func TestIsValidOrderNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid order number",
			number:   "12345678903",
			expected: true,
		},
		{
			name:     "Invalid order number",
			number:   "12345678904",
			expected: false,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidOrderNumber(tt.number)
			if result != tt.expected {
				t.Errorf("IsValidOrderNumber(%s) = %v; want %v", tt.number, result, tt.expected)
			}
		})
	}
}
