package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// User represents a system user account.
type User struct {
	ID        int64    `example:"1"       json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
}

// UserSummary represents a minimal view of a user for listings or references.
type UserSummary struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type password struct {
	text *string
	hash []byte
}

// UserStore handles database operations for users and invitations.
type UserStore struct {
	db *sql.DB
}

func (p *password) Set(text string) error {
	hash, err := p.generateHashFromPassword(text)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) CheckHash(text string) bool {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(text))

	return err == nil
}

func (p *password) generateHashFromPassword(text string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash from password: %w", err)
	}

	return hash, nil
}

// Get fetches a user by their unique identifier.
func (s *UserStore) Get(ctx context.Context, userID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := User{}

	query := `
	SELECT id, username, email, created_at, is_active, password, role_id
	FROM users
	WHERE id = $1 
	`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
		&user.Password.hash,
		&user.RoleID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get user by userID %v: %w", userID, err)
		}
	}

	return &user, nil
}

// CreateAndInvite creates a new user record and generates an invitation token within a transaction.
func (s *UserStore) CreateAndInvite(ctx context.Context, user *UserWithRole, token string, exp time.Duration) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, user.ID, token, exp)
		if err != nil {
			return err
		}

		return nil
	})
}

// ActivateAndClean verifies an invitation token, activates the user account, and removes the token.
func (s *UserStore) ActivateAndClean(ctx context.Context, token string) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		userID, err := s.getUserIDFromInvitationAndCleanIt(ctx, tx, token)
		if err != nil {
			return err
		}

		return s.activateUserByUserID(ctx, tx, userID)
	})
}

// DeleteUser removes a user record by its unique identifier.
func (s *UserStore) DeleteUser(ctx context.Context, userID int64) error {
	query := `
	DELETE FROM users
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return fmt.Errorf("failed to delete user by userID %v: %w", userID, err)
		}
	}

	return nil
}

// DeleteInvitation removes any pending invitation records associated with a user ID.
func (s *UserStore) DeleteInvitation(ctx context.Context, userID int64) error {
	query := `
	DELETE FROM user_invitations
	WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete invitation by userID %v: %w", userID, err)
	}

	return nil
}

// GetByUsername retrieves an active user by their unique username.
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}

	query := `
	SELECT id, username, email, created_at, is_active, password, role_id
	FROM users
	WHERE username = $1 AND is_active = true
	`

	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
		&user.Password.hash,
		&user.RoleID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by username %s: %w", username, err)
	}

	return user, nil
}

// GetByEmail retrieves an active user by their email address.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}

	query := `
	SELECT id, username, email, created_at, is_active, password, role_id
	FROM users
	WHERE email = $1 AND is_active = true
	`

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
		&user.Password.hash,
		&user.RoleID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	return user, nil
}

// DeleteUserAndInvitation deletes both the user record and associated invitations.
func (s *UserStore) DeleteUserAndInvitation(ctx context.Context, userID int64) error {
	err := errors.Join(s.DeleteUser(ctx, userID), s.DeleteInvitation(ctx, userID))

	return err
}

// GetUserWithRole retrieves an active user along with their associated role details.
func (s *UserStore) GetUserWithRole(ctx context.Context, userID int64) (*UserWithRole, error) {
	userWithRole := UserWithRole{}

	query := `
		SELECT
		    u.id,
		    u.username,
		    u.email,
		    u.created_at,
		    u.is_active,
			u.password,

		    r.name,
		    r.description,
		    r.level
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
		  AND u.is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&userWithRole.ID,
		&userWithRole.Username,
		&userWithRole.Email,
		&userWithRole.CreatedAt,
		&userWithRole.IsActive,
		&userWithRole.Password.hash,
		&userWithRole.Role.Name,
		&userWithRole.Role.Description,
		&userWithRole.Role.Level,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get userWithRole by id %d: %w", userID, err)
		}
	}

	return &userWithRole, nil
}

func (s *UserStore) create(ctx context.Context, tx *sql.Tx, userWithRole *UserWithRole) error {
	query := `
	INSERT INTO users (username, email, password, role_id)
	SELECT $1, $2, $3, r.id
	FROM roles r
	WHERE r.name = $4
	RETURNING id, created_at, role_id
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if userWithRole.Role.Name == "" {
		userWithRole.Role.Name = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		userWithRole.Username,
		userWithRole.Email,
		userWithRole.Password.hash,
		userWithRole.Role.Name,
	).Scan(
		&userWithRole.ID,
		&userWithRole.CreatedAt,
		&userWithRole.RoleID,
	)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation code in Postgres
			switch pgErr.Constraint {
			case "users_username_key":
				return ErrDuplicateUsername
			case "users_email_key":
				return ErrDuplicateEmail
			}
		}

		return fmt.Errorf("failed to create user %s: %w", userWithRole.Username, err)
	}

	return nil
}

func (s *UserStore) getUserIDFromInvitationAndCleanIt(ctx context.Context, tx *sql.Tx, token string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
	DELETE FROM user_invitations
	WHERE token = $1 AND expires_at > $2
	RETURNING user_id
	`
	userID := int64(0)

	err := tx.QueryRowContext(
		ctx,
		query,
		[]byte(token),
		time.Now(),
	).Scan(&userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrInvalidToken
		default:
			return 0, fmt.Errorf("failed to consume user invitation: %w", err)
		}
	}

	return userID, nil
}

func (s *UserStore) activateUserByUserID(ctx context.Context, tx *sql.Tx, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
	UPDATE users
	SET is_active = TRUE
	WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to execute user activation query for user %d: %w", userID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for user activation %d: %w", userID, err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, userID int64, token string, exp time.Duration) error {
	query := `
	INSERT INTO user_invitations (token, user_id, expires_at)
	VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		userID,
		time.Now().Add(exp),
	)
	if err != nil {
		return fmt.Errorf("failed to create user invitation for user %d: %w", userID, err)
	}

	return nil
}
