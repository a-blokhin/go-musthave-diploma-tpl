package model

import (
	"strconv"
)

func ValidateLuhn(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	doubleDigits := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if doubleDigits {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		doubleDigits = !doubleDigits
	}

	return sum%10 == 0
}

func IsValidOrderNumber(number string) bool {
	return ValidateLuhn(number)
}
