package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SessionsStore struct {
	db *sql.DB
}

type SessionData struct {
	SessionId string
	UserID    int64

	TokenHash string
	IPAdress  string
	UserAgent string
	IsRevoke  bool

	ExpiresAt time.Time
	CreatedAt string
}

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
		&sessionData.SessionId,
		&sessionData.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

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
		&sessionData.SessionId,
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
			return err
		}
	}

	return nil
}
