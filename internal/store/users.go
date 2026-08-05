package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id" example:"1"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
	INSERT INTO users (username, email, password)
	VALUES ($1, $2, $3) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key" (23505)`:
			return ErrorDuplicateUsername
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key" (23505)`:
			return ErrorDuplicateEmail
		default:
			return err
		}

	}

	return err
}

func (s *UserStore) Get(ctx context.Context, userID int64) (*User, error) {
	user := User{}

	query := `
	SELECT id, username, email, created_at
	FROM users
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
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

func (s *UserStore) ActivateAndClean(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		userID, err := s.getUserIdFromInvitationAndCleanIt(ctx, tx, token)
		if err != nil {
			return err
		}

		return s.activateUserByUserID(ctx, tx, userID)
	})
}

func (s *UserStore) getUserIdFromInvitationAndCleanIt(ctx context.Context, tx *sql.Tx, token string) (int64, error) {
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
			return userID, ErrorInvalidToken
		default:
			return userID, err
		}
	}

	return userID, err
}

func (s *UserStore) activateUserByUserID(ctx context.Context, tx *sql.Tx, userId int64) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
	UPDATE users
	SET is_active = TRUE
	WHERE id = $1
	`
	result, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
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
		return err
	}

	return nil
}

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
		return err
	}

	return nil
}

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
		return err
	}

	return nil
}

func (s *UserStore) DeleteUserAndInvitation(ctx context.Context, userID int64) error {
	err := errors.Join(s.DeleteUser(ctx, userID), s.DeleteInvitation(ctx, userID))

	return err
}
