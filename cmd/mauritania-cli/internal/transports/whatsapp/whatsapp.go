package transports

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// WhatsAppTransport implements the TransportClient interface for WhatsApp using WhatsMeow
type WhatsAppTransport struct {
	client      *whatsmeow.Client
	store       *sqlstore.Container
	config      *utils.Config
	logger      *log.Logger
	rateLimiter *utils.RateLimiter
	messageChan chan *models.IncomingMessage
	isConnected bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWhatsAppTransport creates a new WhatsApp transport client using WhatsMeow
func NewWhatsAppTransport(config *utils.Config, logger *log.Logger) (*WhatsAppTransport, error) {
	ctx, cancel := context.WithCancel(context.Background())

	transport := &WhatsAppTransport{
		config:      config,
		logger:      logger,
		messageChan: make(chan *models.IncomingMessage, 100),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize rate limiter (more lenient for WhatsMeow)
	transport.rateLimiter = utils.NewRateLimiter(1000, time.Hour, logger)

	// Initialize database store for session persistence
	if err := transport.initStore(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	// Create WhatsApp client
	if err := transport.initClient(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return transport, nil
}

// initStore initializes the database store for WhatsApp session persistence
func (w *WhatsAppTransport) initStore() error {
	// Get data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".mauritania-cli", "whatsapp")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize SQLite store
	store, err := sqlstore.New("sqlite3", fmt.Sprintf("file:%s/whatsapp.db?_foreign_keys=on", dataDir), nil)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	w.store = store
	return nil
}

// initClient initializes the WhatsApp client
func (w *WhatsAppTransport) initClient() error {
	deviceStore, err := w.store.GetFirstDevice()
	if err != nil {
		return fmt.Errorf("failed to get device store: %w", err)
	}

	client := whatsmeow.NewWithDevice(deviceStore, nil)
	w.client = client

	// Set up event handlers
	client.AddEventHandler(w.handleEvent)

	return nil
}

// handleEvent handles WhatsApp events
func (w *WhatsAppTransport) handleEvent(evt interface{}) {
	switch e := evt.(type) {
	case *events.Message:
		w.handleMessage(e)
	case *events.QR:
		w.handleQRCode(e)
	case *events.Connected:
		w.logger.Printf("WhatsApp connected as %s", e.ID.String())
		w.isConnected = true
	case *events.Disconnected:
		w.logger.Printf("WhatsApp disconnected")
		w.isConnected = false
	case *events.LoggedOut:
		w.logger.Printf("WhatsApp logged out")
		w.isConnected = false
	}
}

// handleMessage processes incoming WhatsApp messages
func (w *WhatsAppTransport) handleMessage(msg *events.Message) {
	// Skip non-text messages and messages from self
	if msg.Info.IsFromMe || msg.Message.GetConversation() == "" {
		return
	}

	message := &models.IncomingMessage{
		ID:        msg.Info.ID,
		SenderID:  msg.Info.Sender.User,
		Content:   msg.Message.GetConversation(),
		Timestamp: msg.Info.Timestamp,
		Transport: "whatsapp",
	}

	select {
	case w.messageChan <- message:
		w.logger.Printf("Received WhatsApp message from %s: %s", message.SenderID, message.Content[:min(50, len(message.Content))])
	default:
		w.logger.Printf("Message channel full, dropping message from %s", message.SenderID)
	}
}

// handleQRCode displays QR code for authentication
func (w *WhatsAppTransport) handleQRCode(qr *events.QR) {
	w.logger.Printf("WhatsApp QR Code received - scan with your phone")
	qrterminal.GenerateHalfBlock(qr.Codes[0], qrterminal.L, os.Stdout)
}

// SendMessage sends a message via WhatsApp using WhatsMeow
func (w *WhatsAppTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check rate limit
	if w.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Check if connected
	if !w.isConnected {
		return nil, fmt.Errorf("WhatsApp not connected")
	}

	// Parse recipient (should be a phone number)
	recipientJID, err := types.ParseJID(recipient)
	if err != nil {
		// Try to add country code if not present
		if !strings.HasPrefix(recipient, "+") {
			// Assume Mauritania country code (+222) if not specified
			recipientJID, err = types.ParseJID("+222" + recipient)
			if err != nil {
				return nil, fmt.Errorf("invalid recipient format: %s", recipient)
			}
		} else {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
	}

	// Create message
	msg := &proto.Message{
		Conversation: &message,
	}

	// Send message
	resp, err := w.client.SendMessage(w.ctx, recipientJID, msg)
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

// ReceiveMessages retrieves pending messages from the message channel
func (w *WhatsAppTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	var messages []*models.IncomingMessage

	// Collect all available messages without blocking
	for {
		select {
		case msg := <-w.messageChan:
			messages = append(messages, msg)
		default:
			// No more messages available
			break
		}
	}

	return messages, nil
}

// GetStatus returns the current status of the WhatsApp transport
func (w *WhatsAppTransport) GetStatus() (*models.TransportStatus, error) {
	status := &models.TransportStatus{
		Available:   w.isConnected,
		LastChecked: time.Now(),
	}

	if !w.isConnected {
		status.Error = "WhatsApp client not connected"
		if w.client == nil {
			status.Error = "WhatsApp client not initialized"
		}
	} else {
		// Get additional status info
		user, err := w.client.GetUser()
		if err != nil {
			status.Error = fmt.Sprintf("Failed to get user info: %v", err)
		} else {
			w.logger.Printf("WhatsApp connected as: %s", user.String())
		}
	}

	return status, nil
}

// ValidateCredentials validates that the WhatsApp client is properly authenticated
func (w *WhatsAppTransport) ValidateCredentials() error {
	if w.client == nil {
		return fmt.Errorf("WhatsApp client not initialized")
	}

	// Check if we have a stored device (meaning we're logged in)
	device, err := w.store.GetFirstDevice()
	if err != nil {
		return fmt.Errorf("no stored device found - please authenticate first")
	}

	if device == nil {
		return fmt.Errorf("no device registered - please scan QR code first")
	}

	w.logger.Printf("WhatsApp credentials validated - device registered")
	return nil
}

// Connect establishes connection to WhatsApp (non-blocking)
func (w *WhatsAppTransport) Connect() error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	go func() {
		err := w.client.Connect()
		if err != nil {
			w.logger.Printf("Failed to connect to WhatsApp: %v", err)
			return
		}
		w.logger.Printf("WhatsApp connection established")
	}()

	return nil
}

// Disconnect closes the WhatsApp connection
func (w *WhatsAppTransport) Disconnect() error {
	if w.client != nil {
		w.client.Disconnect()
	}
	w.isConnected = false
	w.logger.Printf("WhatsApp disconnected")
	return nil
}

// IsAuthenticated checks if the client is authenticated
func (w *WhatsAppTransport) IsAuthenticated() bool {
	if w.client == nil {
		return false
	}
	return w.client.IsLoggedIn()
}

// GetRateLimit returns current rate limiting status
func (w *WhatsAppTransport) GetRateLimit() (*models.RateLimit, error) {
	return w.rateLimiter.GetStatus()
}

// AuthenticateWithQR initiates QR code authentication (blocking)
func (w *WhatsAppTransport) AuthenticateWithQR() error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	if w.IsAuthenticated() {
		w.logger.Printf("Already authenticated")
		return nil
	}

	w.logger.Printf("Starting WhatsApp authentication with QR code...")
	w.logger.Printf("Please scan the QR code displayed below with your WhatsApp mobile app")

	// Connect and wait for authentication
	err := w.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Wait for authentication to complete
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			w.client.Disconnect()
			return fmt.Errorf("authentication timeout - QR code expired")
		case <-ticker.C:
			if w.IsAuthenticated() {
				w.logger.Printf("WhatsApp authentication successful!")
				return nil
			}
		}
	}
}

// GetQRCode returns the current QR code for authentication
func (w *WhatsAppTransport) GetQRCode() (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	// This would be called from the QR event handler
	// For now, return empty - QR codes are displayed via events
	return "", fmt.Errorf("QR code not available - use AuthenticateWithQR method")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
