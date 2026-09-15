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
	RoleID    int64    `json:"role_id"`
}

type password struct {
	text *string
	hash []byte
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
	if err != nil {
		return false
	}

	return true
}

func (p *password) generateHashFromPassword(text string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
}

type UserStore struct {
	db *sql.DB
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
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *UserWithRole, token string, exp time.Duration) error {
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
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
		return nil, err
	}

	return user, nil
}

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
		return nil, err
	}

	return user, nil
}

func (s *UserStore) DeleteUserAndInvitation(ctx context.Context, userID int64) error {
	err := errors.Join(s.DeleteUser(ctx, userID), s.DeleteInvitation(ctx, userID))

	return err
}

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
			return nil, err
		}
	}

	return &userWithRole, nil
}
