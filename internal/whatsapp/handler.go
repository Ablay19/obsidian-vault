package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for WhatsApp webhooks
type Handler struct {
	service Service
	logger  *otelzap.Logger
}

// NewHandler creates a new WhatsApp handler
func NewHandler(service Service, logger *otelzap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// ServeHTTP implements http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleVerification(w, r)
	case http.MethodPost:
		h.handleWebhook(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleVerification handles WhatsApp webhook verification
func (h *Handler) handleVerification(w http.ResponseWriter, r *http.Request) {
	verifyToken := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	valid, challengeResponse, err := h.service.VerifyWebhook(r.Context(), verifyToken, challenge)
	if err != nil {
		h.logger.Error("Webhook verification error", zap.Error(err))
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	if valid {
		w.WriteHeader(http.StatusOK)
		if challengeResponse != "" {
			w.Write([]byte(challengeResponse))
		}
		h.logger.Info("WhatsApp webhook verified successfully")
	} else {
		w.WriteHeader(http.StatusForbidden)
		h.logger.Warn("WhatsApp webhook verification failed")
	}
}

// handleWebhook handles incoming WhatsApp webhook messages
func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.validateSignature(signature, body) {
		h.logger.Warn("Invalid webhook signature")
		http.Error(w, "Invalid signature", http.StatusForbidden)
		return
	}

	// Parse webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("Failed to unmarshal webhook payload", zap.Error(err))
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Process webhook
	if err := h.service.ProcessWebhook(r.Context(), payload); err != nil {
		h.logger.Error("Failed to process webhook", zap.Error(err))
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("WhatsApp webhook processed successfully")
}

// validateSignature validates the X-Hub-Signature-256 header
func (h *Handler) validateSignature(signature string, payload []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualSignature := strings.TrimPrefix(signature, "sha256=")

	// Get app secret from service configuration
	secret := h.service.GetConfig().AppSecret
	if secret == "" {
		h.logger.Warn("No app secret configured for WhatsApp signature validation")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}

// WhatsAppWebhookHandler is a compatibility function that can be used with the existing system
// This maintains backward compatibility with the current routing setup
func WhatsAppWebhookHandler(service Service, logger *otelzap.Logger) http.HandlerFunc {
	handler := NewHandler(service, logger)
	return handler.ServeHTTP
}
