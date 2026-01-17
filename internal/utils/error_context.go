package utils

import (
	"errors"
	"fmt"
)

// ErrorContext represents structured error context
type ErrorContext struct {
	Service   string
	Operation string
	UserID    string
	Resource  string
	Metadata  map[string]interface{}
}

// ErrorWithContext wraps an error with additional context
type ErrorWithContext struct {
	Err     error
	Context ErrorContext
}

func (e ErrorWithContext) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Context.Service, e.Context.Operation, e.Err.Error())
}

func (e ErrorWithContext) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with context information
func WrapError(err error, service, operation string) error {
	if err == nil {
		return nil
	}

	return ErrorWithContext{
		Err: err,
		Context: ErrorContext{
			Service:   service,
			Operation: operation,
		},
	}
}

// WrapErrorWithUser wraps an error with context including user ID
func WrapErrorWithUser(err error, service, operation, userID string) error {
	if err == nil {
		return nil
	}

	return ErrorWithContext{
		Err: err,
		Context: ErrorContext{
			Service:   service,
			Operation: operation,
			UserID:    userID,
		},
	}
}

// WrapErrorWithResource wraps an error with context including resource information
func WrapErrorWithResource(err error, service, operation, resource string) error {
	if err == nil {
		return nil
	}

	return ErrorWithContext{
		Err: err,
		Context: ErrorContext{
			Service:   service,
			Operation: operation,
			Resource:  resource,
		},
	}
}

// WrapErrorWithContext wraps an error with full context
func WrapErrorWithContext(err error, service, operation, userID, resource string, metadata map[string]interface{}) error {
	if err == nil {
		return nil
	}

	return ErrorWithContext{
		Err: err,
		Context: ErrorContext{
			Service:   service,
			Operation: operation,
			UserID:    userID,
			Resource:  resource,
			Metadata:  metadata,
		},
	}
}

// GetErrorContext extracts context from an error
func GetErrorContext(err error) *ErrorContext {
	var ctxErr ErrorWithContext
	if errors.As(err, &ctxErr) {
		return &ctxErr.Context
	}
	return nil
}

// IsServiceError checks if an error originated from a specific service
func IsServiceError(err error, service string) bool {
	ctx := GetErrorContext(err)
	return ctx != nil && ctx.Service == service
}

// IsOperationError checks if an error occurred during a specific operation
func IsOperationError(err error, operation string) bool {
	ctx := GetErrorContext(err)
	return ctx != nil && ctx.Operation == operation
}

// GetErrorService returns the service that generated the error
func GetErrorService(err error) string {
	ctx := GetErrorContext(err)
	if ctx != nil {
		return ctx.Service
	}
	return ""
}

// GetErrorOperation returns the operation that generated the error
func GetErrorOperation(err error) string {
	ctx := GetErrorContext(err)
	if ctx != nil {
		return ctx.Operation
	}
	return ""
}

// GetErrorUserID returns the user ID associated with the error
func GetErrorUserID(err error) string {
	ctx := GetErrorContext(err)
	if ctx != nil {
		return ctx.UserID
	}
	return ""
}

// GetErrorResource returns the resource associated with the error
func GetErrorResource(err error) string {
	ctx := GetErrorContext(err)
	if ctx != nil {
		return ctx.Resource
	}
	return ""
}

// ErrorChain represents a chain of errors with context
type ErrorChain struct {
	Errors []ErrorWithContext
}

func (ec ErrorChain) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}

	msg := "error chain:"
	for i, err := range ec.Errors {
		msg += fmt.Sprintf("\n  %d. [%s:%s] %s", i+1, err.Context.Service, err.Context.Operation, err.Err.Error())
	}
	return msg
}

// ChainErrors creates an error chain from multiple errors
func ChainErrors(errors ...error) error {
	if len(errors) == 0 {
		return nil
	}

	chain := ErrorChain{}
	for _, err := range errors {
		if ctxErr, ok := err.(ErrorWithContext); ok {
			chain.Errors = append(chain.Errors, ctxErr)
		} else {
			// Wrap plain errors
			chain.Errors = append(chain.Errors, ErrorWithContext{
				Err: err,
				Context: ErrorContext{
					Service:   "unknown",
					Operation: "unknown",
				},
			})
		}
	}

	return chain
}
