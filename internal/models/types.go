package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"password"` // bcrypt hash
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProcessingFile represents a file being processed
type ProcessingFile struct {
	ID          string    `json:"id"`
	Hash        string    `json:"hash"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	ContentType string    `json:"content_type"`
	Status      string    `json:"status"`
	Summary     string    `json:"summary"`
	Topics      []string  `json:"topics"`
	Questions   []string  `json:"questions"`
	AIProvider  string    `json:"ai_provider"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserSession represents a user authentication session
type UserSession struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}
