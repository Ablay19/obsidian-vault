package storage

import (
	"fmt"

	"obsidian-automation/internal/models"
	"obsidian-automation/pkg/utils"
)

// MessageService handles message database operations
type MessageService struct {
	db     *Database
	logger *utils.Logger
}

// NewMessageService creates a new message service
func NewMessageService(db *Database, logger *utils.Logger) *MessageService {
	return &MessageService{
		db:     db,
		logger: logger,
	}
}

// Create creates a new message
func (ms *MessageService) Create(message *models.Message) error {
	query := `
		INSERT INTO messages (conversation_id, user_id, content, message_type, model_used, tokens_used, processing_time, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ms.db.db.Exec(query,
		message.ConversationID,
		message.UserID,
		message.Content,
		message.MessageType,
		message.ModelUsed,
		message.TokensUsed,
		message.ProcessingTime,
		message.CreatedAt,
	)

	if err != nil {
		ms.logger.DatabaseOperation("create", "messages", 0)
		return fmt.Errorf("failed to create message: %w", err)
	}

	ms.logger.DatabaseOperation("create", "messages", 1)
	return nil
}

// GetByConversationID retrieves messages for a conversation with pagination
func (ms *MessageService) GetByConversationID(conversationID int64, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, user_id, content, message_type, model_used, tokens_used, processing_time, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`

	rows, err := ms.db.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.UserID,
			&message.Content,
			&message.MessageType,
			&message.ModelUsed,
			&message.TokensUsed,
			&message.ProcessingTime,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// GetRecentByUserID gets recent messages for a user
func (ms *MessageService) GetRecentByUserID(userID int64, limit int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.user_id, m.content, m.message_type, m.model_used, m.tokens_used, m.processing_time, m.created_at
		FROM messages m
		INNER JOIN conversations c ON m.conversation_id = c.id
		WHERE m.user_id = ?
		ORDER BY m.created_at DESC
		LIMIT ?
	`

	rows, err := ms.db.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.UserID,
			&message.Content,
			&message.MessageType,
			&message.ModelUsed,
			&message.TokensUsed,
			&message.ProcessingTime,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// DeleteConversationMessages removes all messages for a conversation
func (ms *MessageService) DeleteConversationMessages(conversationID int64) error {
	query := `DELETE FROM messages WHERE conversation_id = ?`

	_, err := ms.db.db.Exec(query, conversationID)
	if err != nil {
		return fmt.Errorf("failed to delete conversation messages: %w", err)
	}

	return nil
}

// Count returns total number of messages
func (ms *MessageService) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages`

	err := ms.db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// GetStats returns message service statistics
func (ms *MessageService) GetStats() map[string]interface{} {
	count, err := ms.Count()
	if err != nil {
		return map[string]interface{}{
			"total_messages": 0,
			"error":          err.Error(),
		}
	}

	return map[string]interface{}{
		"total_messages": count,
	}
}
