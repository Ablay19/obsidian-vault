package transports

import (
	"bytes"
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

// TelegramTransport implements the TransportClient interface for Telegram
type TelegramTransport struct {
	config        *utils.Config
	logger        *log.Logger
	httpClient    *http.Client
	botToken      string
	apiURL        string
	allowedChatID string
	rateLimiter   *utils.RateLimiter
	isConnected   bool
}

// NewTelegramTransport creates a new Telegram transport client
func NewTelegramTransport(config *utils.Config, logger *log.Logger) (*TelegramTransport, error) {
	telegramConfig := config.Transports.SocialMedia.Telegram

	transport := &TelegramTransport{
		config:        config,
		logger:        logger,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		botToken:      telegramConfig.BotToken,
		apiURL:        "https://api.telegram.org/bot",
		allowedChatID: telegramConfig.ChatID,
		isConnected:   false,
	}

	// Initialize rate limiter (Telegram allows 30 messages per second)
	transport.rateLimiter = utils.NewRateLimiter(30*60*60, time.Hour, logger) // 30 per second = ~108k per hour

	// Test connection if bot token is configured
	if transport.botToken != "" {
		if err := transport.testConnection(); err != nil {
			transport.logger.Printf("Telegram connection test failed: %v", err)
			return transport, nil
		}
		transport.isConnected = true
		transport.logger.Printf("Telegram transport initialized successfully")
	} else {
		transport.logger.Printf("Telegram bot token not configured")
	}

	return transport, nil
}

// testConnection verifies the Telegram Bot API connection
func (t *TelegramTransport) testConnection() error {
	url := fmt.Sprintf("%s%s/getMe", t.apiURL, t.botToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var telegramResp struct {
		OK     bool `json:"ok"`
		Result struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		} `json:"result"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return fmt.Errorf("Telegram API returned not OK")
	}

	t.logger.Printf("Connected to Telegram bot: @%s (ID: %d)", telegramResp.Result.Username, telegramResp.Result.ID)
	return nil
}

// SendMessage sends a message via Telegram Bot API
func (t *TelegramTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check rate limit
	if t.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Prepare request payload
	requestBody := map[string]interface{}{
		"chat_id":    recipient,
		"text":       message,
		"parse_mode": "Markdown", // Allow basic formatting
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s/sendMessage", t.apiURL, t.botToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := t.httpClient.Do(req)
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
		t.logger.Printf("Telegram API error: %s", string(body))
		return nil, fmt.Errorf("Telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var telegramResp struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			Date int64 `json:"date"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return nil, fmt.Errorf("Telegram API returned not OK")
	}

	// Record rate limit usage
	t.rateLimiter.RecordUsage()

	response := &models.MessageResponse{
		MessageID: strconv.Itoa(telegramResp.Result.MessageID),
		Status:    "sent",
		Timestamp: time.Unix(telegramResp.Result.Date, 0),
	}

	t.logger.Printf("Telegram message sent: %s to chat %d", response.MessageID, telegramResp.Result.Chat.ID)
	return response, nil
}

// ReceiveMessages polls for new messages via Telegram Bot API
func (t *TelegramTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// Get updates from Telegram
	url := fmt.Sprintf("%s%s/getUpdates?timeout=30", t.apiURL, t.botToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	var updatesResp struct {
		OK     bool `json:"ok"`
		Result []struct {
			UpdateID int `json:"update_id"`
			Message  struct {
				MessageID int `json:"message_id"`
				From      struct {
					ID        int64  `json:"id"`
					FirstName string `json:"first_name"`
					Username  string `json:"username"`
				} `json:"from"`
				Chat struct {
					ID   int64  `json:"id"`
					Type string `json:"type"`
				} `json:"chat"`
				Date int64  `json:"date"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &updatesResp); err != nil {
		return nil, fmt.Errorf("failed to parse updates: %w", err)
	}

	var messages []*models.IncomingMessage

	for _, update := range updatesResp.Result {
		if update.Message.Text != "" {
			// Check if message is from allowed chat
			chatIDStr := strconv.FormatInt(update.Message.Chat.ID, 10)
			if t.allowedChatID != "" && chatIDStr != t.allowedChatID {
				t.logger.Printf("Ignoring message from unauthorized chat: %s", chatIDStr)
				continue
			}

			message := &models.IncomingMessage{
				ID:        strconv.Itoa(update.Message.MessageID),
				SenderID:  strconv.FormatInt(update.Message.From.ID, 10),
				Content:   update.Message.Text,
				Timestamp: time.Unix(update.Message.Date, 0),
				Transport: "telegram",
			}

			messages = append(messages, message)
			t.logger.Printf("Received Telegram message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])
		}
	}

	return messages, nil
}

// GetStatus returns the current status of the Telegram transport
func (t *TelegramTransport) GetStatus() (*models.TransportStatus, error) {
	status := &models.TransportStatus{
		Available:   t.isConnected,
		LastChecked: time.Now(),
	}

	if t.botToken == "" {
		status.Error = "Telegram bot token not configured"
		status.Available = false
	} else if !t.isConnected {
		status.Error = "Telegram connection not established"
		// Try to reconnect
		if err := t.testConnection(); err != nil {
			status.Error = err.Error()
		} else {
			t.isConnected = true
			status.Available = true
			status.Error = ""
		}
	}

	return status, nil
}

// ValidateCredentials validates that the Telegram bot credentials are working
func (t *TelegramTransport) ValidateCredentials() error {
	if t.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	// Test the connection
	return t.testConnection()
}

// GetRateLimit returns current rate limiting status
func (t *TelegramTransport) GetRateLimit() (*models.RateLimit, error) {
	return t.rateLimiter.GetStatus()
}

// SetWebhook sets up a webhook for receiving messages (alternative to polling)
func (t *TelegramTransport) SetWebhook(webhookURL string) error {
	requestBody := map[string]interface{}{
		"url": webhookURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s/setWebhook", t.apiURL, t.botToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("setWebhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	var webhookResp struct {
		OK          bool   `json:"ok"`
		Result      bool   `json:"result"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(body, &webhookResp); err != nil {
		return fmt.Errorf("failed to parse webhook response: %w", err)
	}

	if !webhookResp.OK || !webhookResp.Result {
		return fmt.Errorf("setWebhook failed: %s", webhookResp.Description)
	}

	t.logger.Printf("Telegram webhook set to: %s", webhookURL)
	return nil
}

// ProcessWebhook processes incoming Telegram webhook
func (t *TelegramTransport) ProcessWebhook(payload []byte) ([]*models.IncomingMessage, error) {
	var webhookData struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID        int64  `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID   int64  `json:"id"`
				Type string `json:"type"`
			} `json:"chat"`
			Date int64  `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	}

	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	var messages []*models.IncomingMessage

	if webhookData.Message.Text != "" {
		// Check if message is from allowed chat
		chatIDStr := strconv.FormatInt(webhookData.Message.Chat.ID, 10)
		if t.allowedChatID != "" && chatIDStr != t.allowedChatID {
			t.logger.Printf("Ignoring webhook message from unauthorized chat: %s", chatIDStr)
			return messages, nil
		}

		message := &models.IncomingMessage{
			ID:        strconv.Itoa(webhookData.Message.MessageID),
			SenderID:  strconv.FormatInt(webhookData.Message.From.ID, 10),
			Content:   webhookData.Message.Text,
			Timestamp: time.Unix(webhookData.Message.Date, 0),
			Transport: "telegram",
		}

		messages = append(messages, message)
		t.logger.Printf("Processed Telegram webhook message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])
	}

	return messages, nil
}
