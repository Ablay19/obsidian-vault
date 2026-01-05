package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obsidian-automation/internal/pipeline"
	"time"

	"go.uber.org/zap"
)

// MessageType represents the type of WhatsApp message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
)

// Message represents a WhatsApp message
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	Type      MessageType            `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Content   map[string]interface{} `json:"content"`
}

// TextContent represents text message content
type TextContent struct {
	Body string `json:"body"`
}

// MediaContent represents media message content
type MediaContent struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Sha256   string `json:"sha256"`
	FileSize int64  `json:"file_size"`
}

// WebhookPayload represents the WhatsApp webhook payload
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Messages         []Message `json:"messages"`
}

type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// Service defines the WhatsApp service interface
type Service interface {
	VerifyWebhook(ctx context.Context, token, challenge string) (bool, string, error)
	ProcessMessage(ctx context.Context, msg Message) error
	ProcessWebhook(ctx context.Context, payload WebhookPayload) error
	DownloadMedia(ctx context.Context, mediaID string) (*Media, error)
	IsConnected() bool
}

// service implements the WhatsApp service interface
type service struct {
	config       Config
	pipeline     PipelineInterface
	logger       *zap.Logger
	httpClient   *http.Client
	mediaStorage MediaStorage
	validator    Validator
}

// PipelineInterface defines the pipeline interface for dependency injection
type PipelineInterface interface {
	Submit(job pipeline.Job) error
}

// MediaStorage defines the media storage interface
type MediaStorage interface {
	Store(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error)
}

// Validator defines the validation interface
type Validator interface {
	ValidateMessage(msg Message) error
	ValidateMedia(media *Media) error
}

// Config holds the WhatsApp service configuration
type Config struct {
	AccessToken string
	VerifyToken string
	AppSecret   string
	WebhookURL  string
}

// Media represents downloaded media
type Media struct {
	ID           string    `json:"id"`
	MimeType     string    `json:"mime_type"`
	Size         int64     `json:"size"`
	Data         []byte    `json:"data"`
	LocalPath    string    `json:"local_path"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

// NewService creates a new WhatsApp service
func NewService(
	cfg Config,
	pipeline PipelineInterface,
	logger *zap.Logger,
	mediaStorage MediaStorage,
	validator Validator,
) Service {
	return &service{
		config:   cfg,
		pipeline: pipeline,
		logger:   logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		mediaStorage: mediaStorage,
		validator:    validator,
	}
}

// GetConfig returns the service configuration
func (s *service) GetConfig() Config {
	return s.config
}

// VerifyWebhook handles webhook verification
func (s *service) VerifyWebhook(ctx context.Context, token, challenge string) (bool, string, error) {
	if token == s.config.VerifyToken {
		s.logger.Info("WhatsApp webhook verified successfully")
		return true, challenge, nil
	}

	s.logger.Warn("WhatsApp webhook verification failed",
		zap.String("expected_token", s.config.VerifyToken),
		zap.String("received_token", token))
	return false, "", fmt.Errorf("webhook verification failed")
}

// ProcessWebhook processes the entire webhook payload
func (s *service) ProcessWebhook(ctx context.Context, payload WebhookPayload) error {
	if payload.Object != "whatsapp_business_account" {
		return fmt.Errorf("invalid webhook object: %s", payload.Object)
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				for _, msg := range change.Value.Messages {
					if err := s.ProcessMessage(ctx, msg); err != nil {
						s.logger.Error("Failed to process message", zap.Error(err), zap.String("message_id", msg.ID))
						// Continue processing other messages
					}
				}
			}
		}
	}

	return nil
}

// ProcessMessage processes a single WhatsApp message
func (s *service) ProcessMessage(ctx context.Context, msg Message) error {
	// Validate message
	if err := s.validator.ValidateMessage(msg); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	// Process based on message type
	switch msg.Type {
	case MessageTypeText:
		return s.processTextMessage(ctx, msg)
	case MessageTypeImage, MessageTypeDocument, MessageTypeAudio, MessageTypeVideo:
		return s.processMediaMessage(ctx, msg)
	default:
		s.logger.Warn("Unsupported message type", zap.String("type", string(msg.Type)), zap.String("message_id", msg.ID))
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// processTextMessage handles text messages
func (s *service) processTextMessage(ctx context.Context, msg Message) error {
	textContent, ok := msg.Content["text"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid text content in message")
	}

	body, ok := textContent["body"].(string)
	if !ok {
		return fmt.Errorf("invalid text body in message")
	}

	job := pipeline.Job{
		ID:          fmt.Sprintf("wa_%s", msg.ID),
		Source:      "whatsapp",
		SourceID:    msg.ID,
		ContentType: pipeline.ContentTypeText,
		ReceivedAt:  time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   msg.From,
			Language: "English", // TODO: Implement language detection
		},
		Metadata: map[string]interface{}{
			"caption": body,
		},
	}

	return s.pipeline.Submit(job)
}

// processMediaMessage handles media messages
func (s *service) processMediaMessage(ctx context.Context, msg Message) error {
	// Extract media content based on message type
	var mediaContent map[string]interface{}
	switch msg.Type {
	case MessageTypeImage:
		mediaContent, _ = msg.Content["image"].(map[string]interface{})
	case MessageTypeDocument:
		mediaContent, _ = msg.Content["document"].(map[string]interface{})
	case MessageTypeAudio:
		mediaContent, _ = msg.Content["audio"].(map[string]interface{})
	case MessageTypeVideo:
		mediaContent, _ = msg.Content["video"].(map[string]interface{})
	default:
		return fmt.Errorf("unsupported media type: %s", msg.Type)
	}

	mediaID, ok := mediaContent["id"].(string)
	if !ok {
		return fmt.Errorf("media ID not found in message")
	}

	// Download media
	media, err := s.DownloadMedia(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("failed to download media: %w", err)
	}

	// Validate media
	if err := s.validator.ValidateMedia(media); err != nil {
		return fmt.Errorf("media validation failed: %w", err)
	}

	// Determine content type
	contentType := s.mapContentType(media.MimeType)

	job := pipeline.Job{
		ID:            fmt.Sprintf("wa_%s", msg.ID),
		Source:        "whatsapp",
		SourceID:      msg.ID,
		ContentType:   contentType,
		Data:          media.Data,
		FileLocalPath: media.LocalPath,
		ReceivedAt:    time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   msg.From,
			Language: "English", // TODO: Implement language detection
		},
		Metadata: map[string]interface{}{
			"media_id":   media.ID,
			"mime_type":  media.MimeType,
			"file_size":  media.Size,
			"local_path": media.LocalPath,
		},
	}

	return s.pipeline.Submit(job)
}

// DownloadMedia downloads media from WhatsApp API
func (s *service) DownloadMedia(ctx context.Context, mediaID string) (*Media, error) {
	// First, get media metadata
	metadataURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", metadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch media metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("media metadata request failed with status: %d", resp.StatusCode)
	}

	var metadata struct {
		ID       string `json:"id"`
		MimeType string `json:"mime_type"`
		Sha256   string `json:"sha256"`
		FileSize int64  `json:"file_size"`
		URL      string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode media metadata: %w", err)
	}

	// Now download the actual media file
	mediaReq, err := http.NewRequestWithContext(ctx, "GET", metadata.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create media download request: %w", err)
	}

	mediaReq.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	mediaResp, err := s.httpClient.Do(mediaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer mediaResp.Body.Close()

	if mediaResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("media download failed with status: %d", mediaResp.StatusCode)
	}

	data, err := io.ReadAll(mediaResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media data: %w", err)
	}

	// Store media locally
	localPath, err := s.mediaStorage.Store(ctx, mediaID, metadata.MimeType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to store media: %w", err)
	}

	return &Media{
		ID:           metadata.ID,
		MimeType:     metadata.MimeType,
		Size:         metadata.FileSize,
		Data:         data,
		LocalPath:    localPath,
		DownloadedAt: time.Now(),
	}, nil
}

// IsConnected returns the current connection status
func (s *service) IsConnected() bool {
	// TODO: Implement proper connection status checking
	return s.config.AccessToken != "" && s.config.VerifyToken != ""
}

// mapContentType maps MIME types to pipeline content types
func (s *service) mapContentType(mimeType string) pipeline.ContentType {
	switch {
	case mimeType == "application/pdf":
		return pipeline.ContentTypePDF
	case mimeType == "image/jpeg", mimeType == "image/png", mimeType == "image/gif":
		return pipeline.ContentTypeImage
	default:
		return pipeline.ContentTypeText
	}
}
