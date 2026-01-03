package database

import "time"

type ChatMessage struct {
	ID          int
	UserID      int64
	ChatID      int64
	MessageID   int
	Direction   string
	ContentType string
	TextContent string
	FilePath    string
	CreatedAt   time.Time
}

func SaveMessage(userID, chatID int64, messageID int, direction, contentType, text, filePath string) error {
	query := `
        INSERT INTO chat_history 
        (user_id, chat_id, message_id, direction, content_type, text_content, file_path)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := DB.Exec(query, userID, chatID, messageID, direction, contentType, text, filePath)
	return err
}

func GetChatHistory(userID int64, limit int) ([]ChatMessage, error) {
	query := `
        SELECT id, user_id, chat_id, message_id, direction, content_type, text_content, file_path, created_at
        FROM chat_history
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
	rows, err := DB.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.ChatID, &msg.MessageID, &msg.Direction, &msg.ContentType, &msg.TextContent, &msg.FilePath, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
