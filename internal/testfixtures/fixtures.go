package testfixtures

import (
	"time"

	"obsidian-automation/internal/auth"
	"obsidian-automation/internal/vectorstore"
	"obsidian-automation/internal/whatsapp"
)

// TestFixtures provides reusable test data and fixtures
type TestFixtures struct{}

// New creates a new TestFixtures instance
func New() *TestFixtures {
	return &TestFixtures{}
}

// Auth Fixtures

// ValidSession returns a valid session for testing
func (tf *TestFixtures) ValidSession() *auth.Session {
	return &auth.Session{
		ID:        "test_session_123",
		UserID:    "user_456",
		Email:     "test@example.com",
		Name:      "Test User",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// ExpiredSession returns an expired session for testing
func (tf *TestFixtures) ExpiredSession() *auth.Session {
	return &auth.Session{
		ID:        "expired_session_123",
		UserID:    "user_456",
		Email:     "test@example.com",
		Name:      "Test User",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}
}

// ValidOAuthConfig returns valid OAuth configuration
func (tf *TestFixtures) ValidOAuthConfig() auth.OAuthConfig {
	return auth.OAuthConfig{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"openid", "email", "profile"},
	}
}

// ValidSessionConfig returns valid session configuration
func (tf *TestFixtures) ValidSessionConfig() auth.SessionConfig {
	return auth.SessionConfig{
		Secret:     "test-secret-key-that-is-long-enough-for-hmac-123456789",
		Expiration: 24 * time.Hour,
		Secure:     true,
		HTTPOnly:   true,
		CookieName: "test_session",
	}
}

// WhatsApp Fixtures

// ValidWhatsAppConfig returns valid WhatsApp service configuration
func (tf *TestFixtures) ValidWhatsAppConfig() whatsapp.Config {
	return whatsapp.Config{
		AccessToken:   "test_access_token_123",
		VerifyToken:   "test_verify_token_456",
		AppSecret:     "test_app_secret_789",
		WebhookURL:    "https://example.com/webhook",
		PhoneNumberID: "123456789",
	}
}

// TextMessage returns a valid text message for testing
func (tf *TestFixtures) TextMessage() whatsapp.Message {
	return whatsapp.Message{
		ID:   "msg_123",
		From: "1234567890",
		Type: whatsapp.MessageTypeText,
		Content: map[string]interface{}{
			"text": map[string]interface{}{
				"body": "Hello, this is a test message!",
			},
		},
		Timestamp: time.Now().Unix(),
	}
}

// ImageMessage returns a valid image message for testing
func (tf *TestFixtures) ImageMessage() whatsapp.Message {
	return whatsapp.Message{
		ID:   "msg_456",
		From: "1234567890",
		Type: whatsapp.MessageTypeImage,
		Content: map[string]interface{}{
			"image": map[string]interface{}{
				"id":        "img_789",
				"mime_type": "image/jpeg",
				"sha256":    "test_hash_123",
				"file_size": 1024,
				"caption":   "Test image",
			},
		},
		Timestamp: time.Now().Unix(),
	}
}

// AudioMessage returns a valid audio message for testing
func (tf *TestFixtures) AudioMessage() whatsapp.Message {
	return whatsapp.Message{
		ID:   "msg_789",
		From: "1234567890",
		Type: whatsapp.MessageTypeAudio,
		Content: map[string]interface{}{
			"audio": map[string]interface{}{
				"id":        "audio_123",
				"mime_type": "audio/mpeg",
				"sha256":    "audio_hash_456",
				"file_size": 2048,
			},
		},
		Timestamp: time.Now().Unix(),
	}
}

// VideoMessage returns a valid video message for testing
func (tf *TestFixtures) VideoMessage() whatsapp.Message {
	return whatsapp.Message{
		ID:   "msg_999",
		From: "1234567890",
		Type: whatsapp.MessageTypeVideo,
		Content: map[string]interface{}{
			"video": map[string]interface{}{
				"id":        "video_456",
				"mime_type": "video/mp4",
				"sha256":    "video_hash_789",
				"file_size": 10240,
				"caption":   "Test video",
			},
		},
		Timestamp: time.Now().Unix(),
	}
}

// DocumentMessage returns a valid document message for testing
func (tf *TestFixtures) DocumentMessage() whatsapp.Message {
	return whatsapp.Message{
		ID:   "msg_doc_123",
		From: "1234567890",
		Type: whatsapp.MessageTypeDocument,
		Content: map[string]interface{}{
			"document": map[string]interface{}{
				"id":        "doc_789",
				"mime_type": "application/pdf",
				"sha256":    "doc_hash_123",
				"file_size": 5120,
				"filename":  "test_document.pdf",
				"caption":   "Test document",
			},
		},
		Timestamp: time.Now().Unix(),
	}
}

// ValidWebhookPayload returns a valid webhook payload for testing
func (tf *TestFixtures) ValidWebhookPayload() whatsapp.WebhookPayload {
	return whatsapp.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []whatsapp.Entry{
			{
				ID: "test_entry_123",
				Changes: []whatsapp.Change{
					{
						Field: "messages",
						Value: whatsapp.Value{
							MessagingProduct: "whatsapp",
							Metadata: whatsapp.Metadata{
								DisplayPhoneNumber: "1234567890",
								PhoneNumberID:      "123456789",
							},
							Messages: []whatsapp.Message{
								tf.TextMessage(),
								tf.ImageMessage(),
							},
						},
					},
				},
			},
		},
	}
}

// InvalidWebhookPayload returns an invalid webhook payload for testing
func (tf *TestFixtures) InvalidWebhookPayload() whatsapp.WebhookPayload {
	return whatsapp.WebhookPayload{
		Object: "invalid_object",
		Entry:  []whatsapp.Entry{},
	}
}

// Media fixtures

// ValidMedia returns valid media information for testing
func (tf *TestFixtures) ValidMedia() *whatsapp.Media {
	return &whatsapp.Media{
		ID:        "media_123",
		MimeType:  "image/jpeg",
		Sha256:    "test_hash_456",
		FileSize:  2048,
		LocalPath: "/tmp/test_media.jpg",
		URL:       "https://cdn.example.com/media_123.jpg",
	}
}

// LargeMedia returns large media for testing size limits
func (tf *TestFixtures) LargeMedia() *whatsapp.Media {
	return &whatsapp.Media{
		ID:        "large_media_123",
		MimeType:  "video/mp4",
		Sha256:    "large_hash_456",
		FileSize:  100 * 1024 * 1024, // 100MB
		LocalPath: "/tmp/large_video.mp4",
		URL:       "https://cdn.example.com/large_video.mp4",
	}
}

// Vector Store Fixtures

// ValidDocument returns a valid document for vector store testing
func (tf *TestFixtures) ValidDocument() vectorstore.Document {
	return vectorstore.Document{
		ID:      "doc_123",
		Content: "This is a test document for vector store functionality. It contains sample text that can be embedded and searched.",
		Metadata: map[string]interface{}{
			"source":     "test",
			"category":   "documentation",
			"tags":       []string{"test", "vector", "search"},
			"created_at": time.Now(),
			"author":     "test_author",
		},
	}
}

// LargeDocument returns a large document for testing
func (tf *TestFixtures) LargeDocument() vectorstore.Document {
	content := "This is a large document. " + string(make([]byte, 10000)) // 10KB content
	return vectorstore.Document{
		ID:      "large_doc_123",
		Content: content,
		Metadata: map[string]interface{}{
			"source":   "large_test",
			"category": "performance_test",
			"size":     "large",
		},
	}
}

// DocumentBatch returns a batch of documents for testing
func (tf *TestFixtures) DocumentBatch(count int) []vectorstore.Document {
	docs := make([]vectorstore.Document, count)
	for i := 0; i < count; i++ {
		docs[i] = vectorstore.Document{
			ID:      fmt.Sprintf("batch_doc_%d", i),
			Content: fmt.Sprintf("This is batch document number %d with some test content.", i),
			Metadata: map[string]interface{}{
				"batch_id": "test_batch",
				"sequence": i,
				"category": "batch_test",
			},
		}
	}
	return docs
}

// ValidVector returns a valid vector for testing
func (tf *TestFixtures) ValidVector() vectorstore.Vector {
	return vectorstore.Vector{
		ID:     "vec_123",
		Values: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		Metadata: map[string]interface{}{
			"source":      "test",
			"document_id": "doc_123",
			"chunk_index": 0,
		},
		Timestamp: time.Now(),
	}
}

// SearchResult returns a valid search result for testing
func (tf *TestFixtures) SearchResult() vectorstore.SearchResult {
	return vectorstore.SearchResult{
		ID:    "result_123",
		Score: 0.95,
		Metadata: map[string]interface{}{
			"content":     "matching content",
			"document_id": "doc_123",
			"chunk":       0,
		},
	}
}

// SearchResults returns multiple search results for testing
func (tf *TestFixtures) SearchResults(count int) []vectorstore.SearchResult {
	results := make([]vectorstore.SearchResult, count)
	for i := 0; i < count; i++ {
		results[i] = vectorstore.SearchResult{
			ID:    fmt.Sprintf("result_%d", i),
			Score: 1.0 - float64(i)*0.1, // Decreasing scores
			Metadata: map[string]interface{}{
				"content":     fmt.Sprintf("matching content %d", i),
				"document_id": fmt.Sprintf("doc_%d", i),
				"chunk":       i,
			},
		}
	}
	return results
}

// AI Fixtures

// ValidGenerationOptions returns valid AI generation options
func (tf *TestFixtures) ValidGenerationOptions() GenerationOptions {
	return GenerationOptions{
		Model:        "gpt-3.5-turbo",
		MaxTokens:    150,
		Temperature:  0.7,
		TopP:         0.9,
		Stream:       false,
		SystemPrompt: "You are a helpful assistant.",
		UserID:       "user_123",
	}
}

// StreamingGenerationOptions returns options for streaming generation
func (tf *TestFixtures) StreamingGenerationOptions() GenerationOptions {
	opts := tf.ValidGenerationOptions()
	opts.Stream = true
	return opts
}

// LongPrompt returns a long prompt for testing
func (tf *TestFixtures) LongPrompt() string {
	return "Write a comprehensive analysis of artificial intelligence in modern society. " +
		"Cover the following aspects: historical development, current applications, " +
		"ethical considerations, future implications, and societal impact. " +
		"Provide detailed examples and evidence-based reasoning for each point. " +
		"Ensure the analysis is balanced and considers both positive and negative aspects."
}

// ShortPrompt returns a short prompt for testing
func (tf *TestFixtures) ShortPrompt() string {
	return "Explain quantum computing in simple terms."
}

// ValidGenerationResult returns a valid generation result
func (tf *TestFixtures) ValidGenerationResult() GenerationResult {
	return GenerationResult{
		Content:      "This is a sample AI response for testing purposes.",
		Model:        "gpt-3.5-turbo",
		Provider:     "openai",
		UserID:       "user_123",
		InputTokens:  10,
		OutputTokens: 20,
		Cost:         0.002,
		Latency:      2 * time.Second,
		FinishReason: "completed",
	}
}

// ErrorGenerationResult returns a generation result with an error
func (tf *TestFixtures) ErrorGenerationResult() GenerationResult {
	result := tf.ValidGenerationResult()
	result.Content = ""
	result.FinishReason = "error"
	return result
}

// HTTP Request/Response Fixtures

// ValidHTTPRequest returns a valid HTTP request for testing
func (tf *TestFixtures) ValidHTTPRequest() *http.Request {
	req := &http.Request{
		Method: "POST",
		URL:    mustParseURL("https://api.example.com/v1/generate"),
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer test_token"},
			"User-Agent":    []string{"TestClient/1.0"},
		},
		Body: io.NopCloser(strings.NewReader(`{"prompt":"test","model":"gpt-3.5-turbo"}`)),
	}
	return req
}

// ValidHTTPResponse returns a valid HTTP response for testing
func (tf *TestFixtures) ValidHTTPResponse() *http.Response {
	body := `{
		"success": true,
		"data": {
			"content": "Test response",
			"model": "gpt-3.5-turbo",
			"usage": {"input_tokens": 10, "output_tokens": 20}
		}
	}`

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
	return resp
}

// ErrorHTTPResponse returns an error HTTP response for testing
func (tf *TestFixtures) ErrorHTTPResponse() *http.Response {
	body := `{
		"success": false,
		"error": {
			"code": "RATE_LIMIT_EXCEEDED",
			"message": "Rate limit exceeded",
			"details": {"reset_time": 1640995200}
		}
	}`

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Retry-After":  []string{"60"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
	return resp
}

// Database Fixtures

// ValidUser returns a valid user for database testing
func (tf *TestFixtures) ValidUser() User {
	return User{
		ID:        "user_123",
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "$2a$10$test.hashed.password.123456789", // bcrypt hash
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AdminUser returns an admin user for testing
func (tf *TestFixtures) AdminUser() User {
	user := tf.ValidUser()
	user.ID = "admin_123"
	user.Email = "admin@example.com"
	user.Name = "Admin User"
	user.Role = "admin"
	return user
}

// ProcessingFile returns a valid processing file for testing
func (tf *TestFixtures) ProcessingFile() ProcessingFile {
	return ProcessingFile{
		ID:          "file_123",
		Hash:        "abc123def456",
		FileName:    "test_document.pdf",
		FilePath:    "/uploads/test_document.pdf",
		ContentType: "application/pdf",
		Status:      "pending",
		Summary:     "",
		Topics:      []string{},
		Questions:   []string{},
		AIProvider:  "openai",
		UserID:      "user_123",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CompletedProcessingFile returns a completed processing file
func (tf *TestFixtures) CompletedProcessingFile() ProcessingFile {
	file := tf.ProcessingFile()
	file.Status = "completed"
	file.Summary = "This is a test document summary."
	file.Topics = []string{"testing", "documentation", "automation"}
	file.Questions = []string{"What is this document about?", "How does it work?"}
	return file
}

// FailedProcessingFile returns a failed processing file
func (tf *TestFixtures) FailedProcessingFile() ProcessingFile {
	file := tf.ProcessingFile()
	file.Status = "failed"
	return file
}

// Session fixtures

// ValidUserSession returns a valid user session
func (tf *TestFixtures) ValidUserSession() UserSession {
	return UserSession{
		ID:           "session_123",
		UserID:       "user_456",
		SessionToken: "valid_session_token_789",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0 (Test Browser)",
	}
}

// ExpiredUserSession returns an expired user session
func (tf *TestFixtures) ExpiredUserSession() UserSession {
	session := tf.ValidUserSession()
	session.ID = "expired_session_123"
	session.ExpiresAt = time.Now().Add(-1 * time.Hour)
	return session
}

// Utility functions

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

// GetTestData returns test data by key
func (tf *TestFixtures) GetTestData(key string) interface{} {
	testData := map[string]interface{}{
		"valid_user":         tf.ValidUser(),
		"admin_user":         tf.AdminUser(),
		"text_message":       tf.TextMessage(),
		"image_message":      tf.ImageMessage(),
		"audio_message":      tf.AudioMessage(),
		"video_message":      tf.VideoMessage(),
		"document_message":   tf.DocumentMessage(),
		"webhook_payload":    tf.ValidWebhookPayload(),
		"valid_document":     tf.ValidDocument(),
		"large_document":     tf.LargeDocument(),
		"valid_vector":       tf.ValidVector(),
		"search_result":      tf.SearchResult(),
		"generation_options": tf.ValidGenerationOptions(),
		"generation_result":  tf.ValidGenerationResult(),
		"http_request":       tf.ValidHTTPRequest(),
		"http_response":      tf.ValidHTTPResponse(),
		"error_response":     tf.ErrorHTTPResponse(),
		"processing_file":    tf.ProcessingFile(),
		"completed_file":     tf.CompletedProcessingFile(),
		"failed_file":        tf.FailedProcessingFile(),
		"user_session":       tf.ValidUserSession(),
		"expired_session":    tf.ExpiredUserSession(),
		"short_prompt":       tf.ShortPrompt(),
		"long_prompt":        tf.LongPrompt(),
	}

	return testData[key]
}

// GetBatchTestData returns batch test data
func (tf *TestFixtures) GetBatchTestData(key string, count int) interface{} {
	switch key {
	case "documents":
		return tf.DocumentBatch(count)
	case "search_results":
		return tf.SearchResults(count)
	default:
		return nil
	}
}
