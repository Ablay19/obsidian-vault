package transports

import (
	"fmt"
	"log"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// WhatsAppTransport implements the TransportClient interface for WhatsApp
// Note: This is a simplified implementation for basic CLI functionality
// Full WhatsMeow integration requires native Go compilation on target platform
type WhatsAppTransport struct {
	config      *utils.Config
	logger      *log.Logger
	rateLimiter *utils.RateLimiter
	isConnected bool
}

// NewWhatsAppTransport creates a new WhatsApp transport client
func NewWhatsAppTransport(config *utils.Config, logger *log.Logger) (*WhatsAppTransport, error) {
	transport := &WhatsAppTransport{
		config:      config,
		logger:      logger,
		isConnected: false, // Not connected by default
	}

	// Initialize rate limiter (conservative limits for now)
	transport.rateLimiter = utils.NewRateLimiter(100, time.Hour, logger)

	transport.logger.Printf("WhatsApp transport initialized (simplified mode)")
	transport.logger.Printf("Note: Full WhatsApp integration requires WhatsMeow library")
	transport.logger.Printf("Use 'mauritania-cli whatsapp auth' for QR code authentication")

	return transport, nil
}

// SendMessage sends a message via WhatsApp (simplified implementation)
func (w *WhatsAppTransport) SendMessage(recipient, message string) (*models.MessageResponse, error) {
	// Check rate limit
	if w.rateLimiter.IsRateLimited() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// For now, just log the message and return success
	// Full implementation requires WhatsMeow library
	w.logger.Printf("WhatsApp: Would send message to %s: %s", recipient, message[:min(50, len(message))])

	// Record rate limit usage
	w.rateLimiter.RecordUsage()

	response := &models.MessageResponse{
		MessageID: fmt.Sprintf("wa_%d", time.Now().Unix()),
		Status:    "sent",
		Timestamp: time.Now(),
	}

	return response, nil
}

// ReceiveMessages retrieves pending messages (simplified)
func (w *WhatsAppTransport) ReceiveMessages() ([]*models.IncomingMessage, error) {
	// Return empty for now - full implementation needs WhatsMeow
	w.logger.Printf("WhatsApp: ReceiveMessages called (simplified mode)")
	return []*models.IncomingMessage{}, nil
}

// GetStatus returns the current status of the WhatsApp transport
func (w *WhatsAppTransport) GetStatus() (*models.TransportStatus, error) {
	return &models.TransportStatus{
		Available:   w.isConnected,
		LastChecked: time.Now(),
		Error:       "Simplified WhatsApp transport - use 'whatsapp auth' for full setup",
	}, nil
}

// ValidateCredentials validates WhatsApp credentials (simplified)
func (w *WhatsAppTransport) ValidateCredentials() error {
	w.logger.Printf("WhatsApp credentials validation (simplified mode)")
	return fmt.Errorf("simplified WhatsApp transport - full authentication required")
}

// GetRateLimit returns current rate limiting status
func (w *WhatsAppTransport) GetRateLimit() (*models.RateLimit, error) {
	return w.rateLimiter.GetStatus()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
