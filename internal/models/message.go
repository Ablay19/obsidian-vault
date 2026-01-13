package models

import (
	"time"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeUser   MessageType = "user"
	MessageTypeBot    MessageType = "bot"
	MessageTypeSystem MessageType = "system"
)

// Message represents a message in a conversation
type Message struct {
	ID             int64       `json:"id" db:"id"`
	ConversationID int64       `json:"conversation_id" db:"conversation_id"`
	UserID         int64       `json:"user_id" db:"user_id"`
	Content        string      `json:"content" db:"content"`
	MessageType    MessageType `json:"message_type" db:"message_type"`
	ModelUsed      string      `json:"model_used" db:"model_used"`
	TokensUsed     int         `json:"tokens_used" db:"tokens_used"`
	ProcessingTime int         `json:"processing_time" db:"processing_time"` // milliseconds
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// IsFromUser returns true if message is from user
func (m *Message) IsFromUser() bool {
	return m.MessageType == MessageTypeUser
}

// IsFromBot returns true if message is from bot
func (m *Message) IsFromBot() bool {
	return m.MessageType == MessageTypeBot
}

// GetFormattedTimestamp returns a user-friendly timestamp
func (m *Message) GetFormattedTimestamp() string {
	return m.CreatedAt.Format("2006-01-02 15:04:05")
}

// GetShortContent returns shortened version of content
func (m *Message) GetShortContent(maxLength int) string {
	if len(m.Content) <= maxLength {
		return m.Content
	}
	return m.Content[:maxLength-3] + "..."
}
