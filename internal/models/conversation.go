package models

import (
	"time"
)

// Conversation represents a conversation between user and bot
type Conversation struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetDisplayName returns a display name for the conversation
func (c *Conversation) GetDisplayName() string {
	if c.Title != "" {
		return c.Title
	}
	return "New Chat"
}

// IsNew returns true if conversation was created recently (within last hour)
func (c *Conversation) IsNew() bool {
	return time.Since(c.CreatedAt) < time.Hour
}
