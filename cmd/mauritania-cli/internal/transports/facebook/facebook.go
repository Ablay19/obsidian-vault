package transports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// FacebookTransport implements the TransportClient interface for Facebook Messenger
type FacebookTransport struct {
	config      *utils.Config
	logger      *log.Logger
	httpClient  *http.Client
	apiURL      string
	appID       string
	appSecret   string
	accessToken string
	verifyToken string
	rateLimiter *utils.RateLimiter
}

// NewFacebookTransport creates a new Facebook transport client
func NewFacebookTransport(config *utils.Config, logger *log.Logger) (*FacebookTransport, error) {
	facebookConfig := config.Transports.SocialMedia.Facebook

	transport := &FacebookTransport{
		config:      config,
		logger:      logger,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		apiURL:      "https://graph.facebook.com/v18.0",
		appID:       facebookConfig.AppID,
		appSecret:   facebookConfig.AppSecret,
		accessToken: facebookConfig.AccessToken,
		verifyToken: facebookConfig.VerifyToken,
	}

	// Initialize rate limiter (Facebook allows 200 messages per hour)
	transport.rateLimiter = utils.NewRateLimiter(200, time.Hour, logger)

	return transport, nil
}

// SendMessage sends a message via Facebook Messenger API
func (f *FacebookTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check rate limit
	if f.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Prepare request payload
	requestBody := map[string]interface{}{
		"recipient": map[string]string{
			"id": recipient,
		},
		"message": map[string]string{
			"text": message,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/me/messages?access_token=%s", f.apiURL, f.accessToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		f.logger.Printf("Facebook API error: %s", string(body))
		return nil, fmt.Errorf("Facebook API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var facebookResp struct {
		MessageID   string    `json:"message_id"`
		RecipientID string    `json:"recipient_id"`
		Timestamp   time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &facebookResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Record rate limit usage
	f.rateLimiter.RecordUsage()

	response := &models.MessageResponse{
		MessageID: facebookResp.MessageID,
		Status:    "sent",
		Timestamp: time.Now(), // Facebook doesn't return timestamp in response
	}

	f.logger.Printf("Facebook message sent: %s to %s", response.MessageID, recipient)
	return response, nil
}

// ReceiveMessages polls for new messages (Facebook uses webhooks)
func (f *FacebookTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// Facebook primarily uses webhooks for incoming messages
	f.logger.Printf("Facebook transport uses webhooks for receiving messages")
	return []*models.IncomingMessage{}, nil
}

// GetStatus returns the current status of the Facebook transport
func (f *FacebookTransport) GetStatus() (*models.TransportStatus, error) {
	// Test connectivity by calling a simple API endpoint
	url := fmt.Sprintf("%s/me?fields=id,name&access_token=%s", f.apiURL, f.accessToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}, nil
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       "Failed to read response",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var meResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &meResp); err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       "Invalid response format",
		}, nil
	}

	f.logger.Printf("Facebook page authenticated: %s (%s)", meResp.Name, meResp.ID)

	return &models.TransportStatus{
		Available:   true,
		LastChecked: time.Now(),
	}, nil
}

// ValidateCredentials validates that the Facebook credentials are working
func (f *FacebookTransport) ValidateCredentials() error {
	status, err := f.GetStatus()
	if err != nil {
		return err
	}

	if !status.Available {
		return fmt.Errorf("Facebook credentials validation failed: %s", status.Error)
	}

	f.logger.Printf("Facebook credentials validated successfully")
	return nil
}

// GetRateLimit returns current rate limiting status
func (f *FacebookTransport) GetRateLimit() (*models.RateLimit, error) {
	return f.rateLimiter.GetStatus()
}

// ProcessWebhook processes incoming Facebook webhooks
func (f *FacebookTransport) ProcessWebhook(payload []byte) ([]*models.IncomingMessage, error) {
	var webhookData struct {
		Object string `json:"object"`
		Entry  []struct {
			ID      string `json:"id"`
			Changes []struct {
				Field string `json:"field"`
				Value struct {
					Messaging []struct {
						Sender struct {
							ID string `json:"id"`
						} `json:"sender"`
						Recipient struct {
							ID string `json:"id"`
						} `json:"recipient"`
						Timestamp int64 `json:"timestamp"`
						Message   struct {
							MID  string `json:"mid"`
							Text string `json:"text"`
						} `json:"message"`
					} `json:"messaging"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	var messages []*models.IncomingMessage

	for _, entry := range webhookData.Entry {
		for _, change := range entry.Changes {
			for _, messaging := range change.Value.Messaging {
				if messaging.Message.Text != "" {
					message := &models.IncomingMessage{
						ID:        messaging.Message.MID,
						SenderID:  messaging.Sender.ID,
						Content:   messaging.Message.Text,
						Timestamp: time.Unix(messaging.Timestamp/1000, 0), // Convert from milliseconds
						Transport: "facebook",
					}

					messages = append(messages, message)
					f.logger.Printf("Received Facebook message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])
				}
			}
		}
	}

	return messages, nil
}

// VerifyWebhookSignature verifies the webhook signature from Facebook
func (f *FacebookTransport) VerifyWebhookSignature(payload []byte, signature string) error {
	// Facebook webhook signature verification
	// This would implement HMAC-SHA256 verification with app secret
	if f.appSecret == "" {
		return fmt.Errorf("app secret not configured")
	}

	// TODO: Implement proper signature verification
	// For now, just check that signature exists
	if signature == "" {
		return fmt.Errorf("missing webhook signature")
	}

	return nil
}

// SendFile sends a file via Facebook Messenger
func (f *FacebookTransport) SendFile(recipient, filePath string, metadata map[string]interface{}) (*models.FileResponse, error) {
	// Facebook Messenger supports file attachments
	// This would implement file upload to Facebook's servers
	f.logger.Printf("Facebook SendFile not yet implemented: %s to %s", filePath, recipient)

	// Placeholder implementation
	return &models.FileResponse{
		FileID:      "facebook_file_pending",
		FileSize:    0,
		ContentType: "application/octet-stream",
		Status:      "pending_implementation",
		Timestamp:   time.Now(),
	}, fmt.Errorf("SendFile not yet implemented for Facebook transport")
}

// SendBinary sends binary data via Facebook Messenger
func (f *FacebookTransport) SendBinary(recipient string, data []byte, metadata map[string]interface{}) (*models.FileResponse, error) {
	// Facebook Messenger supports binary attachments
	// This would implement binary upload to Facebook's servers
	f.logger.Printf("Facebook SendBinary not yet implemented: %d bytes to %s", len(data), recipient)

	// Placeholder implementation
	return &models.FileResponse{
		FileID:      "facebook_binary_pending",
		FileSize:    int64(len(data)),
		ContentType: "application/octet-stream",
		Status:      "pending_implementation",
		Timestamp:   time.Now(),
	}, fmt.Errorf("SendBinary not yet implemented for Facebook transport")
}
