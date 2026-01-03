package ai

import "fmt"

// AppError represents a custom application error.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Retry   bool   `json:"retry"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Predefined error codes
const (
	ErrCodeInvalidRequest  = 400
	ErrCodeUnauthorized    = 401
	ErrCodeRateLimit       = 429
	ErrCodeInternal        = 500
	ErrCodeProviderOffline = 503
)

// NewError creates a new AppError.
func NewError(code int, message string, err error) *AppError {
	retry := code == ErrCodeRateLimit || code == ErrCodeProviderOffline || code == ErrCodeInternal
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Retry:   retry,
	}
}

// IsRetryable checks if an error should trigger a retry.
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retry
	}
	return false
}
