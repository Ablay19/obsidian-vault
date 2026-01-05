package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// UserRepository handles database operations for users
type UserRepository struct {
	db     *sql.DB
	logger Logger
}

// User represents a user in the database
type User struct {
	ID             string    `db:"id"`
	GoogleID       string    `db:"google_id"`
	Email          string    `db:"email"`
	Name           string    `db:"name"`
	ProfilePicture string    `db:"profile_picture"`
	AccessToken    string    `db:"access_token"`
	RefreshToken   string    `db:"refresh_token"`
	TokenExpiry    time.Time `db:"token_expiry"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// CreateOrUpdateUser creates a new user or updates an existing one
func (ur *UserRepository) CreateOrUpdateUser(ctx context.Context, user *User, tokens *TokenPair) error {
	query := `
		INSERT INTO users (id, google_id, first_name, email, profile_picture, access_token, refresh_token, token_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(google_id) DO UPDATE SET
			first_name = excluded.first_name,
			email = excluded.email,
			profile_picture = excluded.profile_picture,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_expiry = excluded.token_expiry,
			updated_at = CURRENT_TIMESTAMP
	`

	// Truncate ID for compatibility with existing schema
	shortID := user.ID
	if len(shortID) > 10 {
		shortID = shortID[:10]
	}

	result, err := ur.db.ExecContext(ctx, query,
		shortID,
		user.GoogleID,
		user.Name,
		user.Email,
		user.ProfilePicture,
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.Expiry,
	)

	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	ur.logger.Info("User created/updated",
		"google_id", user.GoogleID,
		"email", user.Email,
		"rows_affected", rowsAffected)

	return nil
}

// GetByGoogleID retrieves a user by Google ID
func (ur *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	query := `
		SELECT id, google_id, email, first_name, profile_picture, 
			   access_token, refresh_token, token_expiry, created_at, updated_at
		FROM users 
		WHERE google_id = ?
	`

	var user User
	err := ur.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.ProfilePicture,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with Google ID: %s", googleID)
		}
		return nil, fmt.Errorf("failed to get user by Google ID: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (ur *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, google_id, email, first_name, profile_picture, 
			   access_token, refresh_token, token_expiry, created_at, updated_at
		FROM users 
		WHERE email = ?
	`

	var user User
	err := ur.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.ProfilePicture,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiry,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UpdateTokens updates OAuth tokens for a user
func (ur *UserRepository) UpdateTokens(ctx context.Context, googleID string, tokens *TokenPair) error {
	query := `
		UPDATE users 
		SET access_token = ?, refresh_token = ?, token_expiry = ?, updated_at = CURRENT_TIMESTAMP
		WHERE google_id = ?
	`

	result, err := ur.db.ExecContext(ctx, query,
		tokens.AccessToken,
		tokens.RefreshToken,
		tokens.Expiry,
		googleID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with Google ID: %s", googleID)
	}

	ur.logger.Info("User tokens updated",
		"google_id", googleID,
		"token_expiry", tokens.Expiry)

	return nil
}

// GetTokensForUser retrieves OAuth tokens for a user
func (ur *UserRepository) GetTokensForUser(ctx context.Context, googleID string) (*TokenPair, error) {
	query := `
		SELECT access_token, refresh_token, token_expiry 
		FROM users 
		WHERE google_id = ?
	`

	var accessToken, refreshToken string
	var expiry time.Time

	err := ur.db.QueryRowContext(ctx, query, googleID).Scan(
		&accessToken,
		&refreshToken,
		&expiry,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no tokens found for user: %s", googleID)
		}
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}, nil
}

// DeleteUser removes a user from the database
func (ur *UserRepository) DeleteUser(ctx context.Context, googleID string) error {
	query := `DELETE FROM users WHERE google_id = ?`

	result, err := ur.db.ExecContext(ctx, query, googleID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with Google ID: %s", googleID)
	}

	ur.logger.Info("User deleted", "google_id", googleID)
	return nil
}
