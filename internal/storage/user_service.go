package storage

import (
	"database/sql"
	"fmt"
	"time"

	"obsidian-automation/internal/models"
	"obsidian-automation/pkg/utils"
)

// UserService handles user database operations
type UserService struct {
	db     *sql.DB
	logger *utils.Logger
}

// NewUserService creates a new user service
func NewUserService(db *sql.DB, logger *utils.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (us *UserService) Create(user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, language, personality, preferences, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := us.db.Exec(query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Language,
		user.Personality,
		user.Preferences,
		now,
		now,
	)

	if err != nil {
		us.logger.DatabaseOperation("create", "users", 0)
		return fmt.Errorf("failed to create user: %w", err)
	}

	us.logger.DatabaseOperation("create", "users", 1)
	return nil
}

// GetByID retrieves a user by ID
func (us *UserService) GetByID(id int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language, personality, preferences, created_at, updated_at
		FROM users
		WHERE telegram_id = ?
	`

	user := &models.User{}
	err := us.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Language,
		&user.Personality,
		&user.Preferences,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		us.logger.DatabaseOperation("get", "users", 0)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	us.logger.DatabaseOperation("get", "users", 1)
	return user, nil
}

// Update updates a user
func (us *UserService) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET username = ?, first_name = ?, last_name = ?, language = ?, personality = ?, preferences = ?, updated_at = ?
		WHERE telegram_id = ?
	`

	user.UpdatedAt = time.Now()

	_, err := us.db.Exec(query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Language,
		user.Personality,
		user.Preferences,
		user.UpdatedAt,
		user.TelegramID,
	)

	if err != nil {
		us.logger.DatabaseOperation("update", "users", 0)
		return fmt.Errorf("failed to update user: %w", err)
	}

	us.logger.DatabaseOperation("update", "users", 1)
	return nil
}

// UpdatePreferences updates only user preferences
func (us *UserService) UpdatePreferences(telegramID int64, preferences *models.Preferences) error {
	user, err := us.GetByID(telegramID)
	if err != nil {
		return err
	}

	if err := user.SetPreferences(preferences); err != nil {
		return fmt.Errorf("failed to set preferences: %w", err)
	}

	return us.Update(user)
}

// Delete removes a user
func (us *UserService) Delete(id int64) error {
	query := `DELETE FROM users WHERE telegram_id = ?`

	_, err := us.db.Exec(query, id)
	if err != nil {
		us.logger.DatabaseOperation("delete", "users", 0)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	us.logger.DatabaseOperation("delete", "users", 1)
	return nil
}

// List returns a list of users with pagination
func (us *UserService) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language, personality, preferences, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := us.db.Query(query, limit, offset)
	if err != nil {
		us.logger.DatabaseOperation("list", "users", 0)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Language,
			&user.Personality,
			&user.Preferences,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			us.logger.DatabaseOperation("scan", "users", 0)
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	us.logger.DatabaseOperation("list", "users", 1)
	return users, nil
}

// Count returns total number of users
func (us *UserService) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`

	err := us.db.QueryRow(query).Scan(&count)
	if err != nil {
		us.logger.DatabaseOperation("count", "users", 0)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	us.logger.DatabaseOperation("count", "users", 1)
	return count, nil
}

// Error definitions
var (
	ErrUserNotFound = fmt.Errorf("user not found")
)
