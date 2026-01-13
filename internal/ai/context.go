package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"obsidian-automation/internal/models"
	"obsidian-automation/pkg/utils"
)

// ContextManager manages conversation context and history
type ContextManager struct {
	conversations map[int64]*Conversation
	messages      map[int64][]*models.Message
	logger        *utils.Logger
	mutex         sync.RWMutex
	maxHistory    int
}

// Conversation represents a conversation context
type Conversation struct {
	ID       int64
	UserID   int64
	Messages []*models.Message
	LastUsed time.Time
	mutex    sync.RWMutex
}

// NewContextManager creates a new context manager
func NewContextManager(maxHistory int, logger *utils.Logger) *ContextManager {
	return &ContextManager{
		conversations: make(map[int64]*Conversation),
		messages:      make(map[int64][]*models.Message),
		logger:        logger,
		maxHistory:    maxHistory,
	}
}

// AddMessage adds a message to conversation context
func (cm *ContextManager) AddMessage(conversationID int64, message *models.Message) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Get or create conversation
	conv, exists := cm.conversations[conversationID]
	if !exists {
		conv = &Conversation{
			ID:       conversationID,
			UserID:   message.UserID,
			Messages: []*models.Message{},
		}
		cm.conversations[conversationID] = conv
	}

	conv.mutex.Lock()
	defer conv.mutex.Unlock()

	// Add message to conversation
	conv.Messages = append(conv.Messages, message)
	conv.LastUsed = time.Now()

	// Enforce history limit
	if len(conv.Messages) > cm.maxHistory {
		// Remove oldest messages
		keepCount := len(conv.Messages) - cm.maxHistory
		conv.Messages = conv.Messages[keepCount:]
	}

	cm.logger.Debug("Added message to context",
		"conversation_id", conversationID,
		"message_count", len(conv.Messages),
	)

	// Store in message cache
	cm.messages[conversationID] = conv.Messages
}

// GetContext retrieves conversation context
func (cm *ContextManager) GetContext(conversationID int64, limit int) ([]*models.Message, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	messages, exists := cm.messages[conversationID]
	if !exists {
		return nil, ErrConversationNotFound
	}

	// Return most recent messages (up to limit)
	if limit > 0 && len(messages) > limit {
		start := len(messages) - limit
		return messages[start:], nil
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages for a user
func (cm *ContextManager) GetRecentMessages(userID int64, count int) ([]*models.Message, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var allMessages []*models.Message

	// Collect messages from all conversations for this user
	for _, messages := range cm.messages {
		for _, msg := range messages {
			if msg.UserID == userID {
				allMessages = append(allMessages, msg)
			}
		}
	}

	// Sort by timestamp (most recent first)
	// For simplicity, return in reverse order
	if len(allMessages) > count {
		start := len(allMessages) - count
		return allMessages[start:count], nil
	}

	return allMessages, nil
}

// ClearConversation clears conversation history
func (cm *ContextManager) ClearConversation(conversationID int64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Clear conversation
	delete(cm.conversations, conversationID)
	delete(cm.messages, conversationID)

	cm.logger.Info("Cleared conversation context", "conversation_id", conversationID)
}

// GetConversationStats returns statistics for a conversation
func (cm *ContextManager) GetConversationStats(conversationID int64) map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conv, exists := cm.conversations[conversationID]
	if !exists {
		return map[string]interface{}{
			"message_count": 0,
		}
	}

	conv.mutex.RLock()
	defer conv.mutex.RUnlock()

	stats := map[string]interface{}{
		"message_count":  len(conv.Messages),
		"last_used":      conv.LastUsed,
		"oldest_message": cm.getOldestMessageTime(conv.Messages),
		"newest_message": cm.getNewestMessageTime(conv.Messages),
	}

	return stats
}

// GetUserStats returns statistics for a user
func (cm *ContextManager) GetUserStats(userID int64) map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var userConversations []*Conversation
	var userMessageCount int

	for _, conv := range cm.conversations {
		if conv.UserID == userID {
			userConversations = append(userConversations, conv)
			conv.mutex.RLock()
			userMessageCount += len(conv.Messages)
			conv.mutex.RUnlock()
		}
	}

	stats := map[string]interface{}{
		"conversation_count":  len(userConversations),
		"total_messages":      userMessageCount,
		"active_conversation": cm.getActiveConversation(userConversations),
	}

	return stats
}

// getActiveConversation finds the most recently active conversation
func (cm *ContextManager) getActiveConversation(conversations []*Conversation) map[string]interface{} {
	if len(conversations) == 0 {
		return map[string]interface{}{
			"id":  0,
			"age": 0,
		}
	}

	var mostRecent *Conversation
	for _, conv := range conversations {
		if mostRecent == nil || conv.LastUsed.After(mostRecent.LastUsed) {
			mostRecent = conv
		}
	}

	if mostRecent != nil {
		return map[string]interface{}{
			"id":  mostRecent.ID,
			"age": time.Since(mostRecent.LastUsed).Milliseconds(),
		}
	}

	return map[string]interface{}{
		"id":  0,
		"age": 0,
	}
}

// Helper functions
func (cm *ContextManager) getOldestMessageTime(messages []*models.Message) int64 {
	if len(messages) == 0 {
		return 0
	}

	oldest := messages[0].CreatedAt.Unix()
	for _, msg := range messages {
		if msg.CreatedAt.Unix() < oldest {
			oldest = msg.CreatedAt.Unix()
		}
	}
	return oldest
}

func (cm *ContextManager) getNewestMessageTime(messages []*models.Message) int64 {
	if len(messages) == 0 {
		return 0
	}

	newest := messages[0].CreatedAt.Unix()
	for _, msg := range messages {
		if msg.CreatedAt.Unix() > newest {
			newest = msg.CreatedAt.Unix()
		}
	}
	return newest
}

// Cleanup old contexts
func (cm *ContextManager) Cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Remove conversations older than 24 hours

	for id, conv := range cm.conversations {
		if conv.LastUsed.Before(cutoff) {
			delete(cm.conversations, id)
			delete(cm.messages, id)
			cm.logger.Info("Cleaned up old conversation", "conversation_id", id)
		}
	}
}

// GetStats returns context manager statistics
func (cm *ContextManager) GetStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return map[string]interface{}{
		"total_conversations": len(cm.conversations),
		"total_messages":      len(cm.messages),
		"max_history":         cm.maxHistory,
	}
}

// Error definitions
var (
	ErrConversationNotFound = fmt.Errorf("conversation not found")
)
