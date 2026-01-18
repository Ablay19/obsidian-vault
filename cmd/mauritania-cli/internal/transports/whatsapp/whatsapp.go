package transports

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// WhatsAppTransport implements the TransportClient interface for WhatsApp
// Uses WhatsMeow library for native WhatsApp Web integration
type WhatsAppTransport struct {
	config      *utils.Config
	logger      *log.Logger
	rateLimiter *utils.RateLimiter
	client      *whatsmeow.Client
	storeDir    string
	isConnected bool
}

// NewWhatsAppTransport creates a new WhatsApp transport client using WhatsMeow
func NewWhatsAppTransport(config *utils.Config, logger *log.Logger) (*WhatsAppTransport, error) {
	transport := &WhatsAppTransport{
		config:      config,
		logger:      logger,
		isConnected: false,
	}

	// Initialize rate limiter (WhatsApp: 1000 messages per hour)
	transport.rateLimiter = utils.NewRateLimiter(1000, time.Hour, logger)

	// Set up store directory
	storeDir := config.Transports.SocialMedia.WhatsApp.DatabasePath
	if storeDir == "" {
		homeDir, _ := os.UserHomeDir()
		storeDir = filepath.Join(homeDir, ".mauritania-cli", "whatsapp")
	}

	transport.storeDir = storeDir

	// Create store directory
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	// Initialize database and client
	if err := transport.initClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize WhatsApp client: %w", err)
	}

	transport.logger.Printf("WhatsApp transport initialized with WhatsMeow")
	return transport, nil
}

// initClient initializes the WhatsApp client
func (w *WhatsAppTransport) initClient() error {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	ctx := context.Background()

	container, err := sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s/whatsapp.db?_foreign_keys=on", w.storeDir), dbLog)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			deviceStore = container.NewDevice()
		} else {
			return fmt.Errorf("failed to get device: %w", err)
		}
	}

	logger := waLog.Stdout("Client", "ERROR", true)
	w.client = whatsmeow.NewClient(deviceStore, logger)

	// Check if already authenticated
	if deviceStore.ID != nil {
		w.isConnected = true
		w.logger.Printf("Found existing WhatsApp session")
	}

	return nil
}

// Login initiates the WhatsApp Web login process
func (w *WhatsAppTransport) Login(ctx context.Context) error {
	if w.IsLoggedIn() {
		w.logger.Printf("Already logged in to WhatsApp")
		return nil
	}

	qrChan, err := w.client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	if err := w.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	w.logger.Printf("Scan the QR code below with WhatsApp on your phone:")
	fmt.Println()

	for evt := range qrChan {
		if evt.Event == "code" {
			// Display QR code
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.M, os.Stdout)
			fmt.Println()
			w.logger.Printf("QR code displayed above - scan with WhatsApp")
		} else if evt.Event == "success" {
			w.logger.Printf("✅ Successfully authenticated with WhatsApp!")
			w.isConnected = true
			return nil
		} else if evt.Event == "timeout" {
			return fmt.Errorf("QR code scan timeout - please try again")
		}
	}

	return fmt.Errorf("authentication failed")
}

// IsLoggedIn checks if the client is authenticated
func (w *WhatsAppTransport) IsLoggedIn() bool {
	return w.client.Store.ID != nil
}

// SendMessage sends a message via WhatsApp using WhatsMeow
func (w *WhatsAppTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	ctx := context.Background()

	// Check if connected
	if !w.client.IsConnected() {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	// Check rate limit
	if w.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Parse recipient JID
	recipientJID, err := parseJID(recipient)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient: %w", err)
	}

	// Send message
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	resp, err := w.client.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Record rate limit usage
	w.rateLimiter.RecordUsage()

	response := &models.MessageResponse{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	w.logger.Printf("WhatsApp message sent: %s to %s", response.MessageID, recipient)
	return response, nil
}

// ReceiveMessages retrieves pending messages from WhatsApp
func (w *WhatsAppTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// WhatsMeow handles messages through event handlers
	// For polling, we could check for new messages, but events are preferred
	w.logger.Printf("WhatsApp: ReceiveMessages called (event-driven transport)")
	return []*models.IncomingMessage{}, nil
}

// GetStatus returns the current status of the WhatsApp transport
func (w *WhatsAppTransport) GetStatus() (*models.TransportStatus, error) {
	status := &models.TransportStatus{
		Available:   w.isConnected,
		LastChecked: time.Now(),
	}

	if w.client == nil {
		status.Error = "WhatsApp client not initialized"
		status.Available = false
	} else if !w.IsLoggedIn() {
		status.Error = "WhatsApp not logged in - please run 'mauritania-cli whatsapp login'"
		status.Available = false
	} else if !w.client.IsConnected() {
		status.Error = "WhatsApp client not connected"
		status.Available = false
		w.isConnected = false
	} else {
		status.Error = ""
		w.isConnected = true
	}

	return status, nil
}

// ValidateCredentials validates WhatsApp WhatsMeow connection
func (w *WhatsAppTransport) ValidateCredentials() error {
	if w.client == nil {
		return fmt.Errorf("WhatsApp client not initialized")
	}

	if !w.IsLoggedIn() {
		return fmt.Errorf("WhatsApp not logged in")
	}

	return nil
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

// parseJID parses a JID from a string
func parseJID(jid string) (types.JID, error) {
	if strings.Contains(jid, "@") {
		return types.ParseJID(jid)
	}

	// Assume it's a phone number
	return types.JID{
		User:   jid,
		Server: types.DefaultUserServer,
	}, nil
}
