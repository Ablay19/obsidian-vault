package whatsapp

import (
	"fmt"
	"obsidian-automation/internal/pipeline"
	"strings"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	logger        *otelzap.Logger
	maxFileSize   int64
	allowedTypes  []string
	maxMessageAge time.Duration
}

// ValidatorConfig holds validator configuration
type ValidatorConfig struct {
	MaxFileSize   int64
	AllowedTypes  []string
	MaxMessageAge time.Duration
}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator(config ValidatorConfig, logger *otelzap.Logger) Validator {
	return &DefaultValidator{
		logger:        logger,
		maxFileSize:   config.MaxFileSize,
		allowedTypes:  config.AllowedTypes,
		maxMessageAge: config.MaxMessageAge,
	}
}

// ValidateMessage validates a WhatsApp message
func (v *DefaultValidator) ValidateMessage(msg Message) error {
	// Check required fields
	if msg.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if msg.From == "" {
		return fmt.Errorf("message sender is required")
	}

	if msg.Type == "" {
		return fmt.Errorf("message type is required")
	}

	// Check message age
	if msg.Timestamp > 0 {
		msgTime := time.Unix(msg.Timestamp, 0)
		if time.Since(msgTime) > v.maxMessageAge {
			return fmt.Errorf("message is too old: %v", time.Since(msgTime))
		}
	}

	// Type-specific validation
	switch msg.Type {
	case MessageTypeText:
		return v.validateTextMessage(msg)
	case MessageTypeImage, MessageTypeDocument, MessageTypeAudio, MessageTypeVideo:
		return v.validateMediaMessage(msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// validateTextMessage validates text message content
func (v *DefaultValidator) validateTextMessage(msg Message) error {
	textContent, ok := msg.Content["text"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("text message must have text content")
	}

	body, ok := textContent["body"].(string)
	if !ok {
		return fmt.Errorf("text message must have body")
	}

	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("text message body cannot be empty")
	}

	if len(body) > 4096 {
		return fmt.Errorf("text message body too long: %d characters", len(body))
	}

	return nil
}

// validateMediaMessage validates media message content
func (v *DefaultValidator) validateMediaMessage(msg Message) error {
	// Determine content key based on message type
	var contentKey string
	switch msg.Type {
	case MessageTypeImage:
		contentKey = "image"
	case MessageTypeDocument:
		contentKey = "document"
	case MessageTypeAudio:
		contentKey = "audio"
	case MessageTypeVideo:
		contentKey = "video"
	default:
		return fmt.Errorf("unsupported media type: %s", msg.Type)
	}

	mediaContent, ok := msg.Content[contentKey].(map[string]interface{})
	if !ok {
		return fmt.Errorf("media message must have %s content", contentKey)
	}

	mediaID, ok := mediaContent["id"].(string)
	if !ok {
		return fmt.Errorf("media content must have ID")
	}

	if mediaID == "" {
		return fmt.Errorf("media ID cannot be empty")
	}

	// Validate mime type if present
	if mimeType, ok := mediaContent["mime_type"].(string); ok {
		if !v.isAllowedMimeType(mimeType) {
			return fmt.Errorf("unsupported mime type: %s", mimeType)
		}
	}

	// Validate file size if present
	if fileSize, ok := mediaContent["file_size"].(float64); ok {
		if int64(fileSize) > v.maxFileSize {
			return fmt.Errorf("file size too large: %d bytes", int64(fileSize))
		}
	}

	return nil
}

// ValidateMedia validates downloaded media
func (v *DefaultValidator) ValidateMedia(media *Media) error {
	if media.ID == "" {
		return fmt.Errorf("media ID is required")
	}

	if media.MimeType == "" {
		return fmt.Errorf("media mime type is required")
	}

	if !v.isAllowedMimeType(media.MimeType) {
		return fmt.Errorf("unsupported mime type: %s", media.MimeType)
	}

	if media.Size > v.maxFileSize {
		return fmt.Errorf("media file too large: %d bytes", media.Size)
	}

	if len(media.Data) == 0 && media.LocalPath == "" {
		return fmt.Errorf("media must have either data or local path")
	}

	return nil
}

// isAllowedMimeType checks if mime type is in allowed list
func (v *DefaultValidator) isAllowedMimeType(mimeType string) bool {
	if len(v.allowedTypes) == 0 {
		return true // Allow all if no restrictions
	}

	for _, allowedType := range v.allowedTypes {
		if strings.HasPrefix(mimeType, allowedType) {
			return true
		}
	}

	return false
}

// mapContentType maps MIME types to pipeline content types
func (v *DefaultValidator) MapContentType(mimeType string) pipeline.ContentType {
	switch {
	case mimeType == "application/pdf":
		return pipeline.ContentTypePDF
	case mimeType == "image/jpeg", mimeType == "image/png", mimeType == "image/gif", mimeType == "image/webp":
		return pipeline.ContentTypeImage
	case strings.HasPrefix(mimeType, "text/"):
		return pipeline.ContentTypeText
	default:
		return pipeline.ContentTypeText // Default to text for unknown types
	}
}
