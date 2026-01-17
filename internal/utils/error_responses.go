package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success       bool         `json:"success"`
	Error         ErrorDetails `json:"error"`
	RequestID     string       `json:"request_id,omitempty"`
	Timestamp     time.Time    `json:"timestamp"`
	Path          string       `json:"path,omitempty"`
	Method        string       `json:"method,omitempty"`
	Version       string       `json:"version,omitempty"`
	Documentation string       `json:"documentation,omitempty"`
}

// ErrorDetails contains detailed error information
type ErrorDetails struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// ErrorTemplate defines a reusable error template
type ErrorTemplate struct {
	Code           string
	Title          string
	Description    string
	HTTPStatusCode int
	Suggestions    []string
	Category       string
}

// Predefined error templates
var ErrorTemplates = map[string]ErrorTemplate{
	// Validation errors
	"VALIDATION_REQUIRED": {
		Code:           "VALIDATION_REQUIRED",
		Title:          "Required Field Missing",
		Description:    "A required field is missing from the request",
		HTTPStatusCode: http.StatusBadRequest,
		Suggestions:    []string{"Check the API documentation for required fields", "Ensure all mandatory fields are provided"},
		Category:       "validation",
	},
	"VALIDATION_INVALID_FORMAT": {
		Code:           "VALIDATION_INVALID_FORMAT",
		Title:          "Invalid Field Format",
		Description:    "The provided field value does not match the expected format",
		HTTPStatusCode: http.StatusBadRequest,
		Suggestions:    []string{"Check the field format requirements in the API documentation", "Use the correct data type for the field"},
		Category:       "validation",
	},
	"VALIDATION_TOO_LONG": {
		Code:           "VALIDATION_TOO_LONG",
		Title:          "Field Value Too Long",
		Description:    "The provided field value exceeds the maximum allowed length",
		HTTPStatusCode: http.StatusBadRequest,
		Suggestions:    []string{"Reduce the length of the field value", "Check the maximum allowed length in the API documentation"},
		Category:       "validation",
	},
	"VALIDATION_TOO_SHORT": {
		Code:           "VALIDATION_TOO_SHORT",
		Title:          "Field Value Too Short",
		Description:    "The provided field value is below the minimum required length",
		HTTPStatusCode: http.StatusBadRequest,
		Suggestions:    []string{"Increase the length of the field value", "Check the minimum required length in the API documentation"},
		Category:       "validation",
	},

	// Authentication errors
	"AUTH_INVALID_CREDENTIALS": {
		Code:           "AUTH_INVALID_CREDENTIALS",
		Title:          "Invalid Credentials",
		Description:    "The provided authentication credentials are invalid",
		HTTPStatusCode: http.StatusUnauthorized,
		Suggestions:    []string{"Verify your username/email and password", "Reset your password if you've forgotten it"},
		Category:       "authentication",
	},
	"AUTH_TOKEN_EXPIRED": {
		Code:           "AUTH_TOKEN_EXPIRED",
		Title:          "Authentication Token Expired",
		Description:    "The provided authentication token has expired",
		HTTPStatusCode: http.StatusUnauthorized,
		Suggestions:    []string{"Refresh your authentication token", "Re-authenticate to obtain a new token"},
		Category:       "authentication",
	},
	"AUTH_INSUFFICIENT_PERMISSIONS": {
		Code:           "AUTH_INSUFFICIENT_PERMISSIONS",
		Title:          "Insufficient Permissions",
		Description:    "You do not have sufficient permissions to perform this action",
		HTTPStatusCode: http.StatusForbidden,
		Suggestions:    []string{"Contact your administrator for the required permissions", "Check your account role and permissions"},
		Category:       "authorization",
	},

	// Resource errors
	"RESOURCE_NOT_FOUND": {
		Code:           "RESOURCE_NOT_FOUND",
		Title:          "Resource Not Found",
		Description:    "The requested resource could not be found",
		HTTPStatusCode: http.StatusNotFound,
		Suggestions:    []string{"Verify the resource ID is correct", "Check if the resource has been deleted", "Ensure you have access to the resource"},
		Category:       "resource",
	},
	"RESOURCE_ALREADY_EXISTS": {
		Code:           "RESOURCE_ALREADY_EXISTS",
		Title:          "Resource Already Exists",
		Description:    "A resource with the same identifier already exists",
		HTTPStatusCode: http.StatusConflict,
		Suggestions:    []string{"Use a different identifier", "Check if the resource already exists before creating"},
		Category:       "resource",
	},
	"RESOURCE_IN_USE": {
		Code:           "RESOURCE_IN_USE",
		Title:          "Resource In Use",
		Description:    "The resource is currently in use and cannot be modified or deleted",
		HTTPStatusCode: http.StatusConflict,
		Suggestions:    []string{"Wait for the resource to become available", "Contact the current user of the resource"},
		Category:       "resource",
	},

	// Rate limiting
	"RATE_LIMIT_EXCEEDED": {
		Code:           "RATE_LIMIT_EXCEEDED",
		Title:          "Rate Limit Exceeded",
		Description:    "Too many requests have been made in a short period",
		HTTPStatusCode: http.StatusTooManyRequests,
		Suggestions:    []string{"Wait before making additional requests", "Reduce the frequency of your requests", "Implement exponential backoff"},
		Category:       "rate_limit",
	},

	// Service errors
	"SERVICE_UNAVAILABLE": {
		Code:           "SERVICE_UNAVAILABLE",
		Title:          "Service Unavailable",
		Description:    "The service is temporarily unavailable",
		HTTPStatusCode: http.StatusServiceUnavailable,
		Suggestions:    []string{"Try again later", "Check the service status page", "Contact support if the issue persists"},
		Category:       "service",
	},
	"SERVICE_TIMEOUT": {
		Code:           "SERVICE_TIMEOUT",
		Title:          "Service Timeout",
		Description:    "The request timed out while waiting for a response",
		HTTPStatusCode: http.StatusGatewayTimeout,
		Suggestions:    []string{"Try again with a simpler request", "Check your network connection", "Contact support if the issue persists"},
		Category:       "service",
	},

	// Generic errors
	"INTERNAL_ERROR": {
		Code:           "INTERNAL_ERROR",
		Title:          "Internal Server Error",
		Description:    "An unexpected internal error occurred",
		HTTPStatusCode: http.StatusInternalServerError,
		Suggestions:    []string{"Try again later", "Contact support if the issue persists"},
		Category:       "internal",
	},
	"INVALID_REQUEST": {
		Code:           "INVALID_REQUEST",
		Title:          "Invalid Request",
		Description:    "The request is malformed or contains invalid data",
		HTTPStatusCode: http.StatusBadRequest,
		Suggestions:    []string{"Check the request format and data", "Review the API documentation"},
		Category:       "request",
	},
}

// ErrorResponseBuilder helps build standardized error responses
type ErrorResponseBuilder struct {
	template  *ErrorTemplate
	field     string
	value     interface{}
	details   map[string]interface{}
	requestID string
	path      string
	method    string
	version   string
}

// NewErrorResponseBuilder creates a new error response builder
func NewErrorResponseBuilder(templateKey string) *ErrorResponseBuilder {
	template, exists := ErrorTemplates[templateKey]
	if !exists {
		template = ErrorTemplates["INTERNAL_ERROR"]
	}

	return &ErrorResponseBuilder{
		template: &template,
		details:  make(map[string]interface{}),
	}
}

// WithField sets the field that caused the error
func (erb *ErrorResponseBuilder) WithField(field string) *ErrorResponseBuilder {
	erb.field = field
	return erb
}

// WithValue sets the invalid value
func (erb *ErrorResponseBuilder) WithValue(value interface{}) *ErrorResponseBuilder {
	erb.value = value
	return erb
}

// WithDetail adds additional error details
func (erb *ErrorResponseBuilder) WithDetail(key string, value interface{}) *ErrorResponseBuilder {
	erb.details[key] = value
	return erb
}

// WithRequestID sets the request ID
func (erb *ErrorResponseBuilder) WithRequestID(requestID string) *ErrorResponseBuilder {
	erb.requestID = requestID
	return erb
}

// WithRequestInfo sets request path and method
func (erb *ErrorResponseBuilder) WithRequestInfo(path, method string) *ErrorResponseBuilder {
	erb.path = path
	erb.method = method
	return erb
}

// WithVersion sets the API version
func (erb *ErrorResponseBuilder) WithVersion(version string) *ErrorResponseBuilder {
	erb.version = version
	return erb
}

// Build creates the final error response
func (erb *ErrorResponseBuilder) Build() *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:        erb.template.Code,
			Message:     erb.template.Title,
			Description: erb.template.Description,
			Field:       erb.field,
			Value:       erb.value,
			Details:     erb.details,
			Suggestions: erb.template.Suggestions,
		},
		RequestID: erb.requestID,
		Timestamp: time.Now(),
		Path:      erb.path,
		Method:    erb.method,
		Version:   erb.version,
	}
}

// BuildAndSend builds the error response and sends it as JSON
func (erb *ErrorResponseBuilder) BuildAndSend(w http.ResponseWriter) error {
	response := erb.Build()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(erb.template.HTTPStatusCode)

	return json.NewEncoder(w).Encode(response)
}

// Quick error response functions for common cases
func SendValidationError(w http.ResponseWriter, field string, value interface{}, requestID string) error {
	return NewErrorResponseBuilder("VALIDATION_INVALID_FORMAT").
		WithField(field).
		WithValue(value).
		WithRequestID(requestID).
		BuildAndSend(w)
}

func SendNotFoundError(w http.ResponseWriter, resource string, requestID string) error {
	return NewErrorResponseBuilder("RESOURCE_NOT_FOUND").
		WithDetail("resource", resource).
		WithRequestID(requestID).
		BuildAndSend(w)
}

func SendAuthError(w http.ResponseWriter, requestID string) error {
	return NewErrorResponseBuilder("AUTH_INVALID_CREDENTIALS").
		WithRequestID(requestID).
		BuildAndSend(w)
}

func SendRateLimitError(w http.ResponseWriter, resetTime int64, requestID string) error {
	return NewErrorResponseBuilder("RATE_LIMIT_EXCEEDED").
		WithDetail("reset_time", resetTime).
		WithRequestID(requestID).
		BuildAndSend(w)
}

func SendInternalError(w http.ResponseWriter, requestID string) error {
	return NewErrorResponseBuilder("INTERNAL_ERROR").
		WithRequestID(requestID).
		BuildAndSend(w)
}

// GetErrorTemplate returns an error template by key
func GetErrorTemplate(key string) (*ErrorTemplate, bool) {
	template, exists := ErrorTemplates[key]
	return &template, exists
}

// AddCustomTemplate adds a custom error template
func AddCustomTemplate(key string, template ErrorTemplate) {
	ErrorTemplates[key] = template
}

// GetTemplatesByCategory returns all templates for a specific category
func GetTemplatesByCategory(category string) map[string]ErrorTemplate {
	result := make(map[string]ErrorTemplate)
	for key, template := range ErrorTemplates {
		if template.Category == category {
			result[key] = template
		}
	}
	return result
}
