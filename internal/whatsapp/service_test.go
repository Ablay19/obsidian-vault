package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// Mock implementations for testing
type MockMediaStorage struct {
	mock.Mock
}

func (m *MockMediaStorage) Store(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error) {
	args := m.Called(ctx, mediaID, mimeType, data)
	return args.String(0), args.Error(1)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateMessage(msg Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockValidator) ValidateMedia(media *Media) error {
	args := m.Called(media)
	return args.Error(0)
}

type MockPipelineInterface struct {
	mock.Mock
}

func (m *MockPipelineInterface) Submit(job interface{}) error {
	args := m.Called(job)
	return args.Error(0)
}

func setupTestService(t *testing.T) (*service, *MockMediaStorage, *MockValidator, *MockPipelineInterface) {
	logger := zaptest.NewLogger(t)
	mockStorage := &MockMediaStorage{}
	mockValidator := &MockValidator{}
	mockPipeline := &MockPipelineInterface{}

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	svc := &service{
		config:       config,
		pipeline:     mockPipeline,
		logger:       logger,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		mediaStorage: mockStorage,
		validator:    mockValidator,
	}

	return svc, mockStorage, mockValidator, mockPipeline
}

func TestNewService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockStorage := &MockMediaStorage{}
	mockValidator := &MockValidator{}
	mockPipeline := &MockPipelineInterface{}

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	svc := NewService(config, mockPipeline, logger, mockStorage, mockValidator)

	assert.NotNil(t, svc)
	assert.Equal(t, config, svc.GetConfig())
	assert.NotNil(t, svc.(*service).httpClient)
}

func TestVerifyWebhook_Success(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	verified, challenge, err := svc.VerifyWebhook(context.Background(), "test_verify", "test_challenge")

	assert.NoError(t, err)
	assert.True(t, verified)
	assert.Equal(t, "test_challenge", challenge)
}

func TestVerifyWebhook_Failure(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	verified, challenge, err := svc.VerifyWebhook(context.Background(), "wrong_token", "test_challenge")

	assert.Error(t, err)
	assert.False(t, verified)
	assert.Empty(t, challenge)
	assert.Contains(t, err.Error(), "webhook verification failed")
}

func TestProcessWebhook_ValidPayload(t *testing.T) {
	svc, _, mockValidator, mockPipeline := setupTestService(t)

	payload := WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []Entry{
			{
				ID: "test_entry",
				Changes: []Change{
					{
						Field: "messages",
						Value: Value{
							MessagingProduct: "whatsapp",
							Metadata: Metadata{
								DisplayPhoneNumber: "1234567890",
								PhoneNumberID:      "123456789",
							},
							Messages: []Message{
								{
									ID:   "test_msg_1",
									From: "1234567890",
									Type: MessageTypeText,
									Content: map[string]interface{}{
										"text": map[string]interface{}{
											"body": "Hello World",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Mock validator and pipeline
	mockValidator.On("ValidateMessage", mock.AnythingOfType("Message")).Return(nil)
	mockPipeline.On("Submit", mock.Anything).Return(nil)

	err := svc.ProcessWebhook(context.Background(), payload)

	assert.NoError(t, err)
	mockValidator.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)
}

func TestProcessWebhook_InvalidObject(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	payload := WebhookPayload{
		Object: "invalid_object",
	}

	err := svc.ProcessWebhook(context.Background(), payload)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook object")
}

func TestProcessMessage_TextMessage(t *testing.T) {
	svc, _, mockValidator, mockPipeline := setupTestService(t)

	msg := Message{
		ID:   "test_msg",
		From: "1234567890",
		Type: MessageTypeText,
		Content: map[string]interface{}{
			"text": map[string]interface{}{
				"body": "Hello World",
			},
		},
	}

	// Mock validator and pipeline
	mockValidator.On("ValidateMessage", msg).Return(nil)
	mockPipeline.On("Submit", mock.Anything).Return(nil)

	err := svc.ProcessMessage(context.Background(), msg)

	assert.NoError(t, err)
	mockValidator.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)
}

func TestProcessMessage_ValidationFailure(t *testing.T) {
	svc, _, mockValidator, _ := setupTestService(t)

	msg := Message{
		ID:   "test_msg",
		From: "1234567890",
		Type: MessageTypeText,
	}

	// Mock validator to return error
	mockValidator.On("ValidateMessage", msg).Return(errors.New("validation failed"))

	err := svc.ProcessMessage(context.Background(), msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message validation failed")
	mockValidator.AssertExpectations(t)
}

func TestProcessMessage_UnsupportedType(t *testing.T) {
	svc, _, mockValidator, _ := setupTestService(t)

	msg := Message{
		ID:   "test_msg",
		From: "1234567890",
		Type: "unsupported_type",
	}

	// Mock validator
	mockValidator.On("ValidateMessage", msg).Return(nil)

	err := svc.ProcessMessage(context.Background(), msg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported message type")
	mockValidator.AssertExpectations(t)
}

func TestDownloadMedia_Success(t *testing.T) {
	svc, mockStorage, mockValidator, _ := setupTestService(t)

	// Create test server to mock WhatsApp API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/123456789" {
			// Metadata response
			response := map[string]interface{}{
				"id":        "test_media_id",
				"mime_type": "image/jpeg",
				"sha256":    "test_hash",
				"file_size": 1024,
				"url":       "http://example.com/download",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/download" {
			// Media download response
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("fake image data"))
		}
	}))
	defer server.Close()

	// Update service endpoints to use test server
	svc.(*service).config.AccessToken = "test_token"

	// Override HTTP client to use test server
	svc.(*service).httpClient = server.Client()

	// Mock storage and validator
	mockStorage.On("Store", mock.Anything, "test_media_id", "image/jpeg", []byte("fake image data")).Return("/path/to/media", nil)
	mockValidator.On("ValidateMedia", mock.AnythingOfType("*whatsapp.Media")).Return(nil)

	media, err := svc.DownloadMedia(context.Background(), "test_media_id")

	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, "test_media_id", media.ID)
	assert.Equal(t, "image/jpeg", media.MimeType)
	assert.Equal(t, "/path/to/media", media.LocalPath)

	mockStorage.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestDownloadMedia_MetadataFailure(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc.(*service).httpClient = server.Client()

	_, err := svc.DownloadMedia(context.Background(), "test_media_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media metadata request failed")
}

func TestDownloadMedia_DownloadFailure(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/123456789" {
			// Valid metadata response
			response := map[string]interface{}{
				"id":        "test_media_id",
				"mime_type": "image/jpeg",
				"sha256":    "test_hash",
				"file_size": 1024,
				"url":       "http://example.com/download",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			// Download fails
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc.(*service).httpClient = server.Client()

	_, err := svc.DownloadMedia(context.Background(), "test_media_id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "media download failed")
}

func TestIsConnected(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	// Test with valid config
	assert.True(t, svc.IsConnected())

	// Test with missing access token
	svc.(*service).config.AccessToken = ""
	assert.False(t, svc.IsConnected())

	// Reset for next test
	svc.(*service).config.AccessToken = "test_token"

	// Test with missing phone number ID
	svc.(*service).config.PhoneNumberID = ""
	assert.False(t, svc.IsConnected())
}

func TestMapContentType(t *testing.T) {
	svc, _, _, _ := setupTestService(t)

	tests := []struct {
		mimeType string
		expected string
	}{
		{"application/pdf", "pdf"},
		{"image/jpeg", "image"},
		{"image/png", "image"},
		{"audio/mpeg", "text"},   // fallback
		{"video/mp4", "text"},    // fallback
		{"unknown/type", "text"}, // fallback
	}

	for _, test := range tests {
		result := svc.(*service).mapContentType(test.mimeType)
		assert.Equal(t, test.expected, result.String(), "Failed for MIME type: %s", test.mimeType)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello world", "English"},
		{"مرحبا بالعالم", "Arabic"},
		{"नमस्ते दुनिया", "Hindi"},
		{"Здравствуй мир", "Russian"},
		{"Hola mundo", "Spanish"},
		{"Bonjour le monde", "French"},
		{"Olá mundo", "Portuguese"},
		{"Hallo Welt", "German"},
		{"Merhaba dünya", "Turkish"},
		{"Halo dunia", "Indonesian"},
		{"สวัสดีโลก", "Thai"},
		{"Chào thế giới", "Vietnamese"},
		{"你好世界", "Chinese"},
		{"", "English"},
	}

	for _, test := range tests {
		result := DetectLanguage(test.input)
		assert.Equal(t, test.expected, result, "Failed to detect language for: %s", test.input)
	}
}
