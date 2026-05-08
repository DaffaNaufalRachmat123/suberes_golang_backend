package helpers

import (
	"fmt"
	"unicode"
)

type PasswordValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidatePasswordStrength enforces a strong password policy:
// - Minimum 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
// - At least 1 special character
func ValidatePasswordStrength(password string) PasswordValidationResult {
	var errs []string

	if len(password) < 8 {
		errs = append(errs, "password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errs = append(errs, "password must contain at least 1 uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "password must contain at least 1 lowercase letter")
	}
	if !hasDigit {
		errs = append(errs, "password must contain at least 1 digit")
	}
	if !hasSpecial {
		errs = append(errs, "password must contain at least 1 special character")
	}

	return PasswordValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

// ValidatePasswordStrengthOrError returns a single error message or nil
func ValidatePasswordStrengthOrError(password string) error {
	result := ValidatePasswordStrength(password)
	if result.Valid {
		return nil
	}
	return fmt.Errorf(result.Errors[0])
}
