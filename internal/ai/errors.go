package ai

import "fmt"

// AppError represents a custom application error.
type AppError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	UserMessage string `json:"user_message,omitempty"` // Friendly message for users
	Err         error  `json:"-"`
	Retry       bool   `json:"retry"`
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
	ErrCodePaymentRequired = 402
	ErrCodeForbidden       = 403
	ErrCodeNotFound        = 404
	ErrCodeRateLimit       = 429
	ErrCodeQuotaExceeded   = 402 // Quota exceeded
	ErrCodeInternal        = 500
	ErrCodeProviderOffline = 503
	ErrCodeNetworkError    = 504
	ErrCodeTimeout         = 408
)

// NewError creates a new AppError.
func NewError(code int, message string, err error) *AppError {
	retry := code == ErrCodeRateLimit || code == ErrCodeProviderOffline || code == ErrCodeInternal || code == ErrCodeNetworkError || code == ErrCodeTimeout
	userMsg := getUserFriendlyMessage(code, message)
	return &AppError{
		Code:        code,
		Message:     message,
		UserMessage: userMsg,
		Err:         err,
		Retry:       retry,
	}
}

// getUserFriendlyMessage returns a user-friendly error message based on error code
func getUserFriendlyMessage(code int, technicalMsg string) string {
	switch code {
	case ErrCodeRateLimit:
		return "AI service is busy right now. Please try again in a moment."
	case ErrCodeQuotaExceeded:
		return "AI service usage limit reached. Please try again later."
	case ErrCodeUnauthorized:
		return "Authentication failed with AI service. Please check configuration."
	case ErrCodeProviderOffline:
		return "AI service is temporarily unavailable. Please try again later."
	case ErrCodeNetworkError:
		return "Network connection issue. Please check your internet and try again."
	case ErrCodeTimeout:
		return "AI request timed out. Please try with a shorter message."
	case ErrCodeInvalidRequest:
		return "Invalid request format. Please check your input."
	case ErrCodeInternal:
		return "An internal error occurred. Please try again or contact support."
	default:
		return "AI service encountered an error. Please try again."
	}
}

// IsRetryable checks if an error should trigger a retry.
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retry
	}
	return false
}
