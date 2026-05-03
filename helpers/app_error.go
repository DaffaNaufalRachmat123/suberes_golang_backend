package helpers

import "fmt"

// AppError is a custom error with HTTP code and message
// Use helpers.NewAppError(message, code) to create

type AppError struct {
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(message string, code int) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
	}
}

func (e *AppError) String() string {
	return fmt.Sprintf("AppError: %d %s", e.Code, e.Message)
}
