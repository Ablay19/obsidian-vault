package storage

import (
	"database/sql"
	"fmt"
	"time"

	"obsidian-automation/internal/models"
	"obsidian-automation/pkg/utils"
)

// ConversationService handles conversation database operations
type ConversationService struct {
	db     *Database
	logger *utils.Logger
}

// NewConversationService creates a new conversation service
func NewConversationService(db *Database, logger *utils.Logger) *ConversationService {
	return &ConversationService{
		db:     db,
		logger: logger,
	}
}

// Create creates a new conversation
func (cs *ConversationService) Create(conversation *models.Conversation) error {
	query := `
		INSERT INTO conversations (user_id, title, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	_, err := cs.db.db.Exec(query,
		conversation.UserID,
		conversation.Title,
		now,
		now,
	)

	if err != nil {
		cs.logger.DatabaseOperation("create", "conversations", 0)
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	cs.logger.DatabaseOperation("create", "conversations", 1)
	return nil
}

// GetByID retrieves a conversation by ID
func (cs *ConversationService) GetByID(id int64) (*models.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`

	conversation := &models.Conversation{}
	err := cs.db.db.QueryRow(query, id).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConversationNotFound
		}
		cs.logger.DatabaseOperation("get", "conversations", 0)
		return nil, fmt.Errorf("failed to get conversation by ID: %w", err)
	}

	cs.logger.DatabaseOperation("get", "conversations", 1)
	return conversation, nil
}

// GetByUserID retrieves all conversations for a user
func (cs *ConversationService) GetByUserID(userID int64, limit, offset int) ([]*models.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := cs.db.db.Query(query, userID, limit, offset)
	if err != nil {
		cs.logger.DatabaseOperation("list", "conversations", 0)
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*models.Conversation
	for rows.Next() {
		conversation := &models.Conversation{}
		err := rows.Scan(
			&conversation.ID,
			&conversation.UserID,
			&conversation.Title,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
		)
		if err != nil {
			cs.logger.DatabaseOperation("scan", "conversations", 0)
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}
		conversations = append(conversations, conversation)
	}

	cs.logger.DatabaseOperation("list", "conversations", 1)
	return conversations, nil
}

// Update updates a conversation
func (cs *ConversationService) Update(conversation *models.Conversation) error {
	query := `
		UPDATE conversations 
		SET title = ?, updated_at = ?
		WHERE id = ?
	`

	conversation.UpdatedAt = time.Now()

	_, err := cs.db.db.Exec(query,
		conversation.Title,
		conversation.UpdatedAt,
		conversation.ID,
	)

	if err != nil {
		cs.logger.DatabaseOperation("update", "conversations", 0)
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	cs.logger.DatabaseOperation("update", "conversations", 1)
	return nil
}

// Delete removes a conversation
func (cs *ConversationService) Delete(id int64) error {
	query := `DELETE FROM conversations WHERE id = ?`

	_, err := cs.db.db.Exec(query, id)
	if err != nil {
		cs.logger.DatabaseOperation("delete", "conversations", 0)
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	cs.logger.DatabaseOperation("delete", "conversations", 1)
	return nil
}

// GetLatestByUserID gets the most recent conversation for a user
func (cs *ConversationService) GetLatestByUserID(userID int64) (*models.Conversation, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM conversations
		WHERE user_id = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`

	conversation := &models.Conversation{}
	err := cs.db.db.QueryRow(query, userID).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConversationNotFound
		}
		cs.logger.DatabaseOperation("get", "conversations", 0)
		return nil, fmt.Errorf("failed to get latest conversation: %w", err)
	}

	cs.logger.DatabaseOperation("get", "conversations", 1)
	return conversation, nil
}

// Count returns total number of conversations
func (cs *ConversationService) Count(userID int64) (int, error) {
	var count int
	var query string
	var args []interface{}

	if userID > 0 {
		query = `SELECT COUNT(*) FROM conversations WHERE user_id = ?`
		args = []interface{}{userID}
	} else {
		query = `SELECT COUNT(*) FROM conversations`
		args = []interface{}{}
	}

	err := cs.db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		cs.logger.DatabaseOperation("count", "conversations", 0)
		return 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	cs.logger.DatabaseOperation("count", "conversations", 1)
	return count, nil
}

// Error definitions
var (
	ErrConversationNotFound = fmt.Errorf("conversation not found")
)
