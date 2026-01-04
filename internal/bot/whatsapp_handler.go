package bot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/pipeline"
	"strings"
	"time"

	"go.uber.org/zap"
)

// WhatsAppWebhookHandler handles incoming webhook events from the WhatsApp Business API.
func WhatsAppWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		verifyWebhook(w, r)
		return
	}
	if r.Method == http.MethodPost {
		handleMessage(w, r)
	}
}

func verifyWebhook(w http.ResponseWriter, r *http.Request) {
	verifyToken := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if verifyToken == config.AppConfig.WhatsApp.VerifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		zap.S().Info("WhatsApp webhook verified successfully")
	} else {
		w.WriteHeader(http.StatusForbidden)
		zap.S().Warn("WhatsApp webhook verification failed")
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.S().Error("Failed to read WhatsApp webhook body", "error", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if !validateSignature(r.Header.Get("X-Hub-Signature-256"), body) {
		zap.S().Warn("WhatsApp webhook signature validation failed")
		http.Error(w, "Signature validation failed", http.StatusForbidden)
		return
	}

	var payload WhatsAppWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		zap.S().Error("Failed to unmarshal WhatsApp webhook payload", "error", err)
		http.Error(w, "Failed to unmarshal payload", http.StatusBadRequest)
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				for _, message := range change.Value.Messages {
					processWhatsAppMessage(message)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func validateSignature(signature string, payload []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	actualSignature := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(config.AppConfig.WhatsApp.AppSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}

func processWhatsAppMessage(msg WhatsAppMessage) {
	job := pipeline.Job{
		ID:          fmt.Sprintf("wa_%s", msg.ID),
		Source:      "whatsapp",
		SourceID:    msg.ID,
		ContentType: pipeline.ContentTypeText,
		ReceivedAt:  time.Now(),
		UserContext: pipeline.UserContext{
			UserID:   msg.From,
			Language: "English", // Default, can be improved later
		},
		Metadata: map[string]interface{}{
			"caption": msg.Text.Body,
		},
	}

	// In a real implementation, you would download media here if msg.Type is "image", "document", etc.

	ingestionPipeline.Submit(job)
}

// WhatsAppWebhookPayload represents the top-level structure of a WhatsApp webhook event.
type WhatsAppWebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

type Value struct {
	MessagingProduct string            `json:"messaging_product"`
	Metadata         WhatsAppMetadata  `json:"metadata"`
	Messages         []WhatsAppMessage `json:"messages"`
}

type WhatsAppMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WhatsAppMessage struct {
	From      string       `json:"from"`
	ID        string       `json:"id"`
	Timestamp string       `json:"timestamp"`
	Text      WhatsAppText `json:"text"`
	Type      string       `json:"type"`
}

type WhatsAppText struct {
	Body string `json:"body"`
}
