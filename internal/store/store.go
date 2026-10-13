// Package store provides database interaction models, queries, and persistence methods
// for users, posts, comments, and roles.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// QueryTimeoutDuration defines the standard timeout duration for database queries.
const QueryTimeoutDuration = 5 * time.Second

var (
	// ErrNotFound is returned when a requested database record cannot be found.
	ErrNotFound = errors.New("record not found")
	// ErrVersionConflict is returned when there is a resource version mismatch.
	ErrVersionConflict = errors.New("resource conflict: version mismatch")
	// ErrResourceConflict is returned when a resource already exists.
	ErrResourceConflict = errors.New("resource conflict: already exists")
	// ErrDuplicateEmail is returned when attempting to register an already taken email.
	ErrDuplicateEmail = errors.New("user with this email already exists")
	// ErrDuplicateUsername is returned when attempting to register an already taken username.
	ErrDuplicateUsername = errors.New("user with this username already exists")
	// ErrInvalidToken is returned when an invitation or auth token is invalid or expired.
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrUnexpected is returned for unexpected application or database errors.
	ErrUnexpected = errors.New("unexpected error")
)

// Storage aggregates all data store interfaces for dependency injection.
type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *PostModel) error
		GetPostModel(ctx context.Context, id int64) (*PostModel, error)
		GetPost(ctx context.Context, id int64) (*Post, error)
		Update(ctx context.Context, id int64, req *UpdatePostRequest) error
		Delete(ctx context.Context, id int64) error
		GetUserFeed(ctx context.Context, userID int64, query PaginatedFeedQuery) ([]*PostFeedItem, error)
		GetAllPosts(ctx context.Context, query PaginatedFeedQuery) ([]*PostFeedItem, error)
	}

	Users interface {
		CreateAndInvite(ctx context.Context, user *UserWithRole, token string, invitationDuration time.Duration) error
		ActivateAndClean(ctx context.Context, token string) error
		Get(ctx context.Context, id int64) (*User, error)
		GetByUsername(ctx context.Context, username string) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetUserWithRole(ctx context.Context, id int64) (*UserWithRole, error)
		DeleteInvitation(ctx context.Context, userID int64) error
		DeleteUser(ctx context.Context, id int64) error
		DeleteUserAndInvitation(ctx context.Context, id int64) error
	}

	Comments interface {
		Create(ctx context.Context, comment *Comment) error
		GetByID(ctx context.Context, postID int64, commentID int64) (*Comment, error)
		GetByPostID(ctx context.Context, postID int64) ([]CommentWithUser, error)
		DeleteByPostID(ctx context.Context, postID int64) error
	}

	Followers interface {
		FollowUser(ctx context.Context, followerID, followeeID int64) error
		UnfollowUser(ctx context.Context, followerID, followeeID int64) error
		GetFollowers(ctx context.Context, userID int64) ([]Follower, error)
		GetFollowerModels(ctx context.Context, userID int64) ([]FollowerModel, error)
	}

	Sessions interface {
		CreateSession(ctx context.Context, session *SessionData) error
		GetSessionByTokenHash(ctx context.Context, session *SessionData) error
	}

	Roles interface {
		GetRoleByName(ctx context.Context, name string) (*Role, error)
		GetRoleByID(ctx context.Context, id int64) (*Role, error)
	}
}

// NewStorage initializes and returns a new Storage instance with concrete store implementations.
func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowersStore{db: db},
		Sessions:  &SessionsStore{db: db},
		Roles:     &RolesStore{db: db},
	}
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
