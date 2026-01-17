package transports

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// WhatsAppTransport implements the TransportClient interface for WhatsApp
// Uses WhatsApp Business API for reliable server-side messaging
type WhatsAppTransport struct {
	config        *utils.Config
	logger        *log.Logger
	rateLimiter   *utils.RateLimiter
	httpClient    *http.Client
	apiURL        string
	accessToken   string
	phoneNumberID string
	webhookSecret string
	isConnected   bool
	lastUpdateID  int64
}

// NewWhatsAppTransport creates a new WhatsApp transport client using Business API
func NewWhatsAppTransport(config *utils.Config, logger *log.Logger) (*WhatsAppTransport, error) {
	transport := &WhatsAppTransport{
		config:       config,
		logger:       logger,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiURL:       "https://graph.facebook.com/v18.0/",
		isConnected:  false,
		lastUpdateID: 0,
	}

	// Get configuration values
	whatsappConfig := config.Transports.SocialMedia.WhatsApp
	transport.accessToken = config.Transports.Shipper.APIKey      // Use shipper API key as access token for now
	transport.phoneNumberID = config.Transports.Shipper.APISecret // Use shipper API secret as phone number ID for now
	transport.webhookSecret = whatsappConfig.WebhookSecret

	// Initialize rate limiter (WhatsApp Business API: 250 messages per day for free tier)
	transport.rateLimiter = utils.NewRateLimiter(250, 24*time.Hour, logger)

	// Check if properly configured
	if transport.accessToken == "" || transport.phoneNumberID == "" {
		transport.logger.Printf("WhatsApp Business API not configured - access token and phone number ID required")
		return transport, nil
	}

	// Test connection
	if err := transport.testConnection(); err != nil {
		transport.logger.Printf("WhatsApp connection test failed: %v", err)
		return transport, nil
	}

	transport.isConnected = true
	transport.logger.Printf("WhatsApp Business API transport initialized successfully")
	return transport, nil
}

// testConnection verifies the WhatsApp Business API connection
func (w *WhatsAppTransport) testConnection() error {
	url := fmt.Sprintf("%s%s", w.apiURL, w.phoneNumberID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.accessToken)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	w.logger.Printf("WhatsApp Business API connection test successful")
	return nil
}

// SendMessage sends a message via WhatsApp Business API
func (w *WhatsAppTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check if configured
	if w.accessToken == "" || w.phoneNumberID == "" {
		return nil, fmt.Errorf("WhatsApp Business API not configured - missing access token or phone number ID")
	}

	// Check rate limit
	if w.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Prepare request payload
	requestBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s/messages", w.apiURL, w.phoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+w.accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := w.httpClient.Do(req)
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
		w.logger.Printf("WhatsApp API error: %s", string(body))
		return nil, fmt.Errorf("WhatsApp API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var whatsappResp struct {
		MessagingProduct string `json:"messaging_product"`
		Contacts         []struct {
			Input string `json:"input"`
			WaID  string `json:"wa_id"`
		} `json:"contacts"`
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(body, &whatsappResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(whatsappResp.Messages) == 0 {
		return nil, fmt.Errorf("no message ID returned from WhatsApp API")
	}

	// Record rate limit usage
	w.rateLimiter.RecordUsage()

	response := &models.MessageResponse{
		MessageID: whatsappResp.Messages[0].ID,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	w.logger.Printf("WhatsApp message sent: %s to %s", response.MessageID, recipient)
	return response, nil
}

// ReceiveMessages retrieves pending messages via webhook or polling
func (w *WhatsAppTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// WhatsApp Business API primarily uses webhooks, but we can poll for messages
	// For now, return empty as webhooks are the preferred method
	w.logger.Printf("WhatsApp: ReceiveMessages called (webhook-based transport)")
	return []*models.IncomingMessage{}, nil
}

// GetStatus returns the current status of the WhatsApp transport
func (w *WhatsAppTransport) GetStatus() (*models.TransportStatus, error) {
	status := &models.TransportStatus{
		Available:   w.isConnected,
		LastChecked: time.Now(),
	}

	if w.accessToken == "" || w.phoneNumberID == "" {
		status.Error = "WhatsApp Business API not configured - missing access token or phone number ID"
		status.Available = false
	} else if !w.isConnected {
		status.Error = "WhatsApp Business API connection not established"
	} else {
		status.Error = ""
	}

	return status, nil
}

// ValidateCredentials validates WhatsApp Business API credentials
func (w *WhatsAppTransport) ValidateCredentials() error {
	if w.accessToken == "" {
		return fmt.Errorf("WhatsApp access token not configured")
	}

	if w.phoneNumberID == "" {
		return fmt.Errorf("WhatsApp phone number ID not configured")
	}

	// Test the connection
	return w.testConnection()
}

// GetRateLimit returns current rate limiting status
func (w *WhatsAppTransport) GetRateLimit() (*models.RateLimit, error) {
	return w.rateLimiter.GetStatus()
}

// ProcessWebhook processes incoming WhatsApp webhook
func (w *WhatsAppTransport) ProcessWebhook(payload []byte, signature string) ([]*models.IncomingMessage, error) {
	// Verify webhook signature first
	if err := w.VerifyWebhookSignature(payload, signature); err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	var webhookData struct {
		Object string `json:"object"`
		Entry  []struct {
			ID      string `json:"id"`
			Changes []struct {
				Value struct {
					MessagingProduct string `json:"messaging_product"`
					Metadata         struct {
						DisplayPhoneNumber string `json:"display_phone_number"`
						PhoneNumberID      string `json:"phone_number_id"`
					} `json:"metadata"`
					Contacts []struct {
						Profile struct {
							Name string `json:"name"`
						} `json:"profile"`
						WaID string `json:"wa_id"`
					} `json:"contacts"`
					Messages []struct {
						ID        string `json:"id"`
						From      string `json:"from"`
						Timestamp string `json:"timestamp"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
						Type string `json:"type"`
					} `json:"messages"`
				} `json:"value"`
				Field string `json:"field"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	var messages []*models.IncomingMessage

	for _, entry := range webhookData.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				for _, msg := range change.Value.Messages {
					if msg.Type == "text" && msg.Text.Body != "" {
						// Parse timestamp
						timestamp, err := strconv.ParseInt(msg.Timestamp, 10, 64)
						if err != nil {
							w.logger.Printf("Failed to parse timestamp %s: %v", msg.Timestamp, err)
							timestamp = time.Now().Unix()
						}

						message := &models.IncomingMessage{
							ID:        msg.ID,
							SenderID:  msg.From,
							Content:   msg.Text.Body,
							Timestamp: time.Unix(timestamp, 0),
							Transport: "whatsapp",
						}

						messages = append(messages, message)
						w.logger.Printf("Processed WhatsApp webhook message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])
					}
				}
			}
		}
	}

	return messages, nil
}

// VerifyWebhookSignature verifies the webhook signature from WhatsApp using HMAC-SHA256
func (w *WhatsAppTransport) VerifyWebhookSignature(payload []byte, signature string) error {
	// Get webhook secret from config
	webhookSecret := w.config.Transports.SocialMedia.WhatsApp.WebhookSecret
	if webhookSecret == "" {
		return fmt.Errorf("WhatsApp webhook secret not configured")
	}

	// WhatsApp sends signature in format "sha256=<signature>"
	if len(signature) <= 7 || signature[:7] != "sha256=" {
		return fmt.Errorf("invalid signature format")
	}
	providedSignature := signature[7:]

	// Calculate expected signature using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return fmt.Errorf("webhook signature verification failed")
	}

	w.logger.Printf("WhatsApp webhook signature verified successfully")
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
