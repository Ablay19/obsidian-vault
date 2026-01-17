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

// WhatsAppTransport implements the TransportClient interface for WhatsApp
type WhatsAppTransport struct {
	config        *utils.Config
	logger        *log.Logger
	httpClient    *http.Client
	apiURL        string
	accessToken   string
	phoneNumberID string
	webhookToken  string
	rateLimiter   *utils.RateLimiter
}

// NewWhatsAppTransport creates a new WhatsApp transport client
func NewWhatsAppTransport(config *utils.Config, logger *log.Logger) (*WhatsAppTransport, error) {
	// Get WhatsApp config from main config
	whatsappConfig := config.Transports.SocialMedia.WhatsApp

	transport := &WhatsAppTransport{
		config:        config,
		logger:        logger,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		apiURL:        "https://graph.facebook.com/v18.0",
		accessToken:   whatsappConfig.APIKey,
		phoneNumberID: whatsappConfig.PhoneNumberID,
		webhookToken:  whatsappConfig.WebhookToken,
	}

	// Initialize rate limiter (WhatsApp allows 250 messages per hour)
	transport.rateLimiter = utils.NewRateLimiter(250, time.Hour, logger)

	return transport, nil
}

// SendMessage sends a message via WhatsApp
func (w *WhatsAppTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check rate limit
	if w.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Prepare WhatsApp API request
	requestBody := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
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
	url := fmt.Sprintf("%s/%s/messages", w.apiURL, w.phoneNumberID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
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
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
		Contacts []struct {
			Input string `json:"input"`
			WAID  string `json:"wa_id"`
		} `json:"contacts"`
	}

	if err := json.Unmarshal(body, &whatsappResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(whatsappResp.Messages) == 0 {
		return nil, fmt.Errorf("no message ID in response")
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

// ReceiveMessages polls for new messages (WhatsApp uses webhooks, so this may not be needed)
func (w *WhatsAppTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// WhatsApp primarily uses webhooks for incoming messages
	// This method could be used for polling as a fallback
	w.logger.Printf("WhatsApp transport uses webhooks for receiving messages")
	return []*models.IncomingMessage{}, nil
}

// GetStatus returns the current status of the WhatsApp transport
func (w *WhatsAppTransport) GetStatus() (*models.TransportStatus, error) {
	// Test connectivity by making a simple API call
	url := fmt.Sprintf("%s/%s", w.apiURL, w.phoneNumberID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}, nil
	}

	req.Header.Set("Authorization", "Bearer "+w.accessToken)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return &models.TransportStatus{
			Available:   false,
			LastChecked: time.Now(),
			Error:       err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	available := resp.StatusCode == http.StatusOK
	var errorMsg string
	if !available {
		errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return &models.TransportStatus{
		Available:   available,
		LastChecked: time.Now(),
		Error:       errorMsg,
	}, nil
}

// ValidateCredentials validates that the WhatsApp credentials are working
func (w *WhatsAppTransport) ValidateCredentials() error {
	status, err := w.GetStatus()
	if err != nil {
		return err
	}

	if !status.Available {
		return fmt.Errorf("WhatsApp credentials validation failed: %s", status.Error)
	}

	w.logger.Printf("WhatsApp credentials validated successfully")
	return nil
}

// GetRateLimit returns current rate limiting status
func (w *WhatsAppTransport) GetRateLimit() (*models.RateLimit, error) {
	return w.rateLimiter.GetStatus()
}

// ProcessWebhook processes incoming WhatsApp webhooks
func (w *WhatsAppTransport) ProcessWebhook(payload []byte) ([]*models.IncomingMessage, error) {
	var webhookData struct {
		Object string `json:"object"`
		Entry  []struct {
			Changes []struct {
				Value struct {
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
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	var messages []*models.IncomingMessage

	for _, entry := range webhookData.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				if msg.Type == "text" && msg.Text.Body != "" {
					// Parse timestamp
					timestamp := time.Now() // Default to now
					if ts, err := time.Parse("0", msg.Timestamp); err == nil {
						timestamp = ts
					}

					message := &models.IncomingMessage{
						ID:        msg.ID,
						SenderID:  msg.From,
						Content:   msg.Text.Body,
						Timestamp: timestamp,
						Transport: "whatsapp",
					}

					messages = append(messages, message)
					w.logger.Printf("Received WhatsApp message from %s: %s", msg.From, msg.Text.Body[:min(50, len(msg.Text.Body))])
				}
			}
		}
	}

	return messages, nil
}

// VerifyWebhookSignature verifies the webhook signature from WhatsApp
func (w *WhatsAppTransport) VerifyWebhookSignature(payload []byte, signature string) error {
	// WhatsApp webhook signature verification
	// This would implement HMAC-SHA256 verification with the webhook token
	if w.webhookToken == "" {
		return fmt.Errorf("webhook token not configured")
	}

	// TODO: Implement proper signature verification
	// For now, just check that signature exists
	if signature == "" {
		return fmt.Errorf("missing webhook signature")
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
