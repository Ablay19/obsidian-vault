package models

import (
	"encoding/json"
	"time"
)

// User represents a bot user
type User struct {
	ID          int64     `json:"id" db:"id"`
	TelegramID  int64     `json:"telegram_id" db:"telegram_id"`
	Username    string    `json:"username" db:"username"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Language    string    `json:"language" db:"language"`
	Personality string    `json:"personality" db:"personality"`
	Preferences string    `json:"preferences" db:"preferences"` // JSON string
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Preferences represents user preferences as a structured type
type Preferences struct {
	Notifications bool   `json:"notifications"`
	ResponseStyle string `json:"response_style"`
	Language      string `json:"language"`
	Model         string `json:"model"`
}

// GetPreferences returns user preferences as structured data
func (u *User) GetPreferences() (*Preferences, error) {
	if u.Preferences == "" {
		return &Preferences{
			Notifications: true,
			ResponseStyle: "helpful",
			Language:      "en",
			Model:         "local",
		}, nil
	}

	var prefs Preferences
	if err := json.Unmarshal([]byte(u.Preferences), &prefs); err != nil {
		return nil, err
	}

	return &prefs, nil
}

// SetPreferences sets user preferences from structured data
func (u *User) SetPreferences(prefs *Preferences) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	u.Preferences = string(data)
	return nil
}
