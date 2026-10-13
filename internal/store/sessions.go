package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SessionsStore handles database operations for user sessions.
type SessionsStore struct {
	db *sql.DB
}

// SessionData holds the fields associated with an active or stored user session.
type SessionData struct {
	SessionID string
	UserID    int64

	TokenHash string
	Token     string `json:"-"`

	IPAdress  string
	UserAgent string
	IsRevoke  bool

	ExpiresAt time.Time
	CreatedAt string
}

// CreateSession inserts a new user session record into the database and returns its assigned ID and creation timestamp.
func (db *SessionsStore) CreateSession(ctx context.Context, sessionData *SessionData) error {
	query := `
		INSERT INTO users_sessions (user_id, token_hash, user_agent, ip_address, is_revoked, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := db.db.QueryRowContext(ctx,
		query,
		sessionData.UserID,
		sessionData.TokenHash,
		sessionData.UserAgent,
		sessionData.IPAdress,
		sessionData.IsRevoke,
		sessionData.ExpiresAt,
	).Scan(
		&sessionData.SessionID,
		&sessionData.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session for user %d: %w", sessionData.UserID, err)
	}

	return nil
}

// GetSessionByTokenHash retrieves a session record using its unique token hash.
func (db *SessionsStore) GetSessionByTokenHash(ctx context.Context, sessionData *SessionData) error {
	query := `
	SELECT id, user_id, token_hash, user_agent, ip_address, is_revoked, expires_at, created_at
	FROM users_sessions 
	WHERE token_hash = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := db.db.QueryRowContext(ctx,
		query,
		sessionData.TokenHash,
	).Scan(
		&sessionData.SessionID,
		&sessionData.UserID,

		&sessionData.TokenHash,
		&sessionData.UserAgent,
		&sessionData.IPAdress,
		&sessionData.IsRevoke,

		&sessionData.ExpiresAt,
		&sessionData.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return fmt.Errorf("failed to get session by tokenHash %s: %w", sessionData.TokenHash, err)
		}
	}

	return nil
}
