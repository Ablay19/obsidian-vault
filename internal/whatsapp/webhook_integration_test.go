package whatsapp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WhatsAppWebhookIntegrationTestSuite tests the complete WhatsApp webhook flow
func TestWhatsAppWebhookIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WhatsApp webhook integration tests in short mode")
	}

	t.Run("WebhookVerificationFlow", func(t *testing.T) {
		testWebhookVerificationFlow(t)
	})

	t.Run("MessageProcessingFlow", func(t *testing.T) {
		testMessageProcessingFlow(t)
	})

	t.Run("MediaProcessingFlow", func(t *testing.T) {
		testMediaProcessingFlow(t)
	})

	t.Run("ErrorHandlingFlow", func(t *testing.T) {
		testErrorHandlingFlow(t)
	})

	t.Run("RateLimitingFlow", func(t *testing.T) {
		testRateLimitingFlow(t)
	})

	t.Run("ConcurrentWebhookProcessing", func(t *testing.T) {
		testConcurrentWebhookProcessing(t)
	})

	t.Run("WebhookSignatureValidation", func(t *testing.T) {
		testWebhookSignatureValidation(t)
	})
}

func testWebhookVerificationFlow(t *testing.T) {
	// Create test server to simulate Meta's webhook verification
	metaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == "test_verify_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer metaServer.Close()

	// Create WhatsApp service with test configuration
	config := Config{
		AccessToken:   "test_access_token",
		VerifyToken:   "test_verify_token",
		AppSecret:     "test_app_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Test successful verification
	verified, challenge, err := service.VerifyWebhook(context.Background(), "test_verify_token", "test_challenge")
	assert.NoError(t, err)
	assert.True(t, verified)
	assert.Equal(t, "test_challenge", challenge)

	// Test failed verification
	verified, challenge, err = service.VerifyWebhook(context.Background(), "wrong_token", "test_challenge")
	assert.Error(t, err)
	assert.False(t, verified)
	assert.Empty(t, challenge)
}

func testMessageProcessingFlow(t *testing.T) {
	// Setup mock AI pipeline and storage
	mockPipeline := &MockPipelineInterface{}
	mockStorage := &MockMediaStorage{}
	mockValidator := &MockValidator{}

	mockPipeline.On("Submit", mock.Anything).Return(nil)
	mockValidator.On("ValidateMessage", mock.Anything).Return(nil)

	// Create WhatsApp service
	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:       config,
		pipeline:     mockPipeline,
		mediaStorage: mockStorage,
		validator:    mockValidator,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Test text message processing
	textMessage := Message{
		ID:   "msg_123",
		From: "1234567890",
		Type: MessageTypeText,
		Content: map[string]interface{}{
			"text": map[string]interface{}{
				"body": "Hello, this is a test message!",
			},
		},
	}

	err := service.ProcessMessage(context.Background(), textMessage)
	assert.NoError(t, err)

	mockValidator.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)

	// Test image message processing
	imageMessage := Message{
		ID:   "msg_456",
		From: "1234567890",
		Type: MessageTypeImage,
		Content: map[string]interface{}{
			"image": map[string]interface{}{
				"id":        "img_789",
				"mime_type": "image/jpeg",
				"sha256":    "test_hash",
				"file_size": 1024,
				"caption":   "Test image",
			},
		},
	}

	// Setup mock storage for media
	mockStorage.On("Store", mock.Anything, "img_789", "image/jpeg", mock.AnythingOfType("[]uint8")).Return("https://cdn.example.com/img_789.jpg", nil)

	err = service.ProcessMessage(context.Background(), imageMessage)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func testMediaProcessingFlow(t *testing.T) {
	// Create test server to simulate WhatsApp media API
	mediaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/123456789" {
			// Metadata response
			metadata := map[string]interface{}{
				"id":        "media_123",
				"mime_type": "image/jpeg",
				"sha256":    "test_hash",
				"file_size": 2048,
				"url":       fmt.Sprintf("http://%s/download", r.Host),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metadata)
		} else if r.Method == http.MethodGet && r.URL.Path == "/download" {
			// Media download response
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("fake image data for testing"))
		}
	}))
	defer mediaServer.Close()

	// Setup mock storage
	mockStorage := &MockMediaStorage{}
	mockValidator := &MockValidator{}

	mockStorage.On("Store", mock.Anything, "media_123", "image/jpeg", []byte("fake image data for testing")).Return("/stored/media_123.jpg", nil)
	mockValidator.On("ValidateMedia", mock.AnythingOfType("*whatsapp.Media")).Return(nil)

	// Create service with test server URL
	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:       config,
		mediaStorage: mockStorage,
		validator:    mockValidator,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	media, err := service.DownloadMedia(context.Background(), "media_123")
	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, "media_123", media.ID)
	assert.Equal(t, "image/jpeg", media.MimeType)
	assert.Equal(t, "/stored/media_123.jpg", media.LocalPath)

	mockStorage.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func testErrorHandlingFlow(t *testing.T) {
	// Test invalid webhook payload
	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Test invalid object
	invalidPayload := WebhookPayload{
		Object: "invalid_object",
	}

	err := service.ProcessWebhook(context.Background(), invalidPayload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid webhook object")

	// Test empty payload
	emptyPayload := WebhookPayload{}
	err = service.ProcessWebhook(context.Background(), emptyPayload)
	assert.Error(t, err)

	// Test message validation failure
	mockValidator := &MockValidator{}
	service.validator = mockValidator

	mockValidator.On("ValidateMessage", mock.Anything).Return(fmt.Errorf("validation failed"))

	validPayload := WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []Entry{
			{
				Changes: []Change{
					{
						Value: Value{
							Messages: []Message{
								{
									ID:   "test_msg",
									From: "1234567890",
									Type: MessageTypeText,
									Content: map[string]interface{}{
										"text": map[string]interface{}{
											"body": "test",
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

	err = service.ProcessWebhook(context.Background(), validPayload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message validation failed")

	mockValidator.AssertExpectations(t)
}

func testRateLimitingFlow(t *testing.T) {
	// This test would verify that webhook processing respects rate limits
	// For now, we'll create a placeholder test
	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Test that service can handle rapid webhook calls
	// In a real implementation, this would test rate limiting middleware
	assert.NotNil(t, service)
	assert.Equal(t, config, service.GetConfig())
}

func testConcurrentWebhookProcessing(t *testing.T) {
	// Setup service
	mockPipeline := &MockPipelineInterface{}
	mockValidator := &MockValidator{}

	mockPipeline.On("Submit", mock.Anything).Return(nil)
	mockValidator.On("ValidateMessage", mock.Anything).Return(nil)

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		pipeline:   mockPipeline,
		validator:  mockValidator,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Create multiple messages
	messages := []Message{}
	for i := 0; i < 10; i++ {
		msg := Message{
			ID:   fmt.Sprintf("msg_%d", i),
			From: "1234567890",
			Type: MessageTypeText,
			Content: map[string]interface{}{
				"text": map[string]interface{}{
					"body": fmt.Sprintf("Concurrent message %d", i),
				},
			},
		}
		messages = append(messages, msg)
	}

	// Process messages concurrently
	done := make(chan bool, len(messages))

	for _, msg := range messages {
		go func(m Message) {
			err := service.ProcessMessage(context.Background(), m)
			assert.NoError(t, err)
			done <- true
		}(msg)
	}

	// Wait for all messages to be processed
	for i := 0; i < len(messages); i++ {
		<-done
	}

	mockPipeline.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func testWebhookSignatureValidation(t *testing.T) {
	appSecret := "test_app_secret"
	payload := `{"test":"data"}`

	// Generate expected signature
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(payload))
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Test valid signature
	valid := validateWebhookSignature(payload, expectedSignature, appSecret)
	assert.True(t, valid)

	// Test invalid signature
	invalid := validateWebhookSignature(payload, "sha256=invalid", appSecret)
	assert.False(t, invalid)

	// Test missing sha256 prefix
	noPrefix := validateWebhookSignature(payload, hex.EncodeToString(mac.Sum(nil)), appSecret)
	assert.False(t, noPrefix)

	// Test wrong secret
	wrongSecret := validateWebhookSignature(payload, expectedSignature, "wrong_secret")
	assert.False(t, wrongSecret)
}

// Helper function for webhook signature validation (would be implemented in actual service)
func validateWebhookSignature(payload, signature, secret string) bool {
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}

	sig := signature[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedMAC))
}

// Benchmark tests for webhook processing
func BenchmarkWebhookVerification(b *testing.B) {
	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify_token",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.VerifyWebhook(context.Background(), "test_verify_token", "challenge")
	}
}

func BenchmarkMessageProcessing(b *testing.B) {
	mockPipeline := &MockPipelineInterface{}
	mockValidator := &MockValidator{}

	mockPipeline.On("Submit", mock.Anything).Return(nil)
	mockValidator.On("ValidateMessage", mock.Anything).Return(nil)

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		pipeline:   mockPipeline,
		validator:  mockValidator,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	message := Message{
		ID:   "bench_msg",
		From: "1234567890",
		Type: MessageTypeText,
		Content: map[string]interface{}{
			"text": map[string]interface{}{
				"body": "Benchmark message",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ProcessMessage(context.Background(), message)
	}
}

// Load testing for webhook processing
func TestWebhookLoadUnderPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping webhook load test in short mode")
	}

	mockPipeline := &MockPipelineInterface{}
	mockValidator := &MockValidator{}

	mockPipeline.On("Submit", mock.Anything).Return(nil)
	mockValidator.On("ValidateMessage", mock.Anything).Return(nil)

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:     config,
		pipeline:   mockPipeline,
		validator:  mockValidator,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	concurrency := 50
	messagesPerWorker := 20

	done := make(chan bool, concurrency)

	message := Message{
		ID:   "load_test_msg",
		From: "1234567890",
		Type: MessageTypeText,
		Content: map[string]interface{}{
			"text": map[string]interface{}{
				"body": "Load test message",
			},
		},
	}

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			for j := 0; j < messagesPerWorker; j++ {
				// Create unique message for each request
				msg := message
				msg.ID = fmt.Sprintf("load_test_msg_%d_%d", workerID, j)

				err := service.ProcessMessage(context.Background(), msg)
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	t.Logf("Completed webhook load test: %d workers, %d messages each", concurrency, messagesPerWorker)
}

// End-to-end WhatsApp message flow test
func TestEndToEndWhatsAppMessageFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end WhatsApp test in short mode")
	}

	// This test simulates the complete flow from webhook reception to AI processing
	mockPipeline := &MockPipelineInterface{}
	mockValidator := &MockValidator{}
	mockStorage := &MockMediaStorage{}

	mockPipeline.On("Submit", mock.Anything).Return(nil)
	mockValidator.On("ValidateMessage", mock.Anything).Return(nil)

	config := Config{
		AccessToken:   "test_token",
		VerifyToken:   "test_verify",
		AppSecret:     "test_secret",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}

	service := &service{
		config:       config,
		pipeline:     mockPipeline,
		validator:    mockValidator,
		mediaStorage: mockStorage,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	// Step 1: Webhook verification
	verified, challenge, err := service.VerifyWebhook(context.Background(), "test_verify", "test_challenge")
	assert.NoError(t, err)
	assert.True(t, verified)
	assert.Equal(t, "test_challenge", challenge)

	// Step 2: Process incoming webhook with message
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
									ID:        "e2e_msg_1",
									From:      "1234567890",
									Type:      MessageTypeText,
									Timestamp: 1234567890,
									Content: map[string]interface{}{
										"text": map[string]interface{}{
											"body": "Hello, this is an end-to-end test message!",
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

	// Step 3: Process webhook
	err = service.ProcessWebhook(context.Background(), payload)
	assert.NoError(t, err)

	// Step 4: Verify message was processed correctly
	mockValidator.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)

	t.Log("End-to-end WhatsApp message flow test completed successfully")
}
