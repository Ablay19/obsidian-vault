package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type QueuedMessage struct {
	ID        string            `json:"id"`
	JID       string            `json:"jid"`
	Text      string            `json:"text"`
	Media     *MediaAttachment  `json:"media,omitempty"`
	Priority  int               `json:"priority"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type MediaAttachment struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
	Caption  string `json:"caption,omitempty"`
}

func handleOutgoingMessage(d amqp091.Delivery) {
	var msg QueuedMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("Failed to unmarshal queued message: %v", err)
		return
	}

	log.Printf("Processing outgoing message to %s: %s", msg.JID, msg.Text)

	// Load session if needed
	if wac == nil {
		conn, err := loadSession()
		if err != nil {
			log.Printf("Failed to load session for queued message: %v", err)
			// Requeue with delay
			time.Sleep(5 * time.Second)
			d.Nack(false, true)
			return
		}
		wac = conn
	}

	// Send message
	if msg.Media != nil {
		// Handle media message
		err := sendMediaMessage(msg.JID, msg.Media)
		if err != nil {
			log.Printf("Failed to send media message: %v", err)
			d.Nack(false, true)
			return
		}
	} else {
		// Handle text message
		textMsg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: msg.JID,
			},
			Text: msg.Text,
		}

		_, err := wac.Send(textMsg)
		if err != nil {
			log.Printf("Failed to send queued message: %v", err)
			d.Nack(false, true)
			return
		}
	}

	log.Printf("Successfully sent queued message to %s", msg.JID)
}

func handleAIResponse(d amqp091.Delivery) {
	var response struct {
		JID      string `json:"jid"`
		Query    string `json:"query"`
		Response string `json:"response"`
	}

	if err := json.Unmarshal(d.Body, &response); err != nil {
		log.Printf("Failed to unmarshal AI response: %v", err)
		return
	}

	log.Printf("Received AI response for %s", response.JID)

	// Send AI response back to WhatsApp
	if wac == nil {
		conn, err := loadSession()
		if err != nil {
			log.Printf("Failed to load session for AI response: %v", err)
			return
		}
		wac = conn
	}

	reply := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: response.JID,
		},
		Text: fmt.Sprintf("🤖 %s", response.Response),
	}

	_, err := wac.Send(reply)
	if err != nil {
		log.Printf("Failed to send AI response: %v", err)
		return
	}

	log.Printf("Sent AI response to %s", response.JID)
}

func handleSystemEvent(d amqp091.Delivery) {
	event := string(d.Body)
	log.Printf("System event: %s", event)

	// Handle system events like connection status, errors, etc.
	// Could trigger notifications, restarts, etc.
}

func queueMessage(jid, text string, priority int) error {
	if queueMgr == nil {
		return fmt.Errorf("queue manager not initialized")
	}

	msg := QueuedMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		JID:       jid,
		Text:      text,
		Priority:  priority,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("%s.%s", config.RabbitMQ.Queues.Outgoing, jid)
	return queueMgr.PublishMessage(routingKey, string(data))
}

func sendMediaMessage(jid string, media *MediaAttachment) error {
	// Implementation for sending media messages
	// This would use whatsapp.ImageMessage, DocumentMessage, etc.
	log.Printf("Media sending not yet implemented for %s: %s", jid, media.Filename)
	return nil
}
