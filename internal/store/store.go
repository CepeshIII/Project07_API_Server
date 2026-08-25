package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetByID(context.Context, int64) (*Post, error)
		Update(context.Context, int64, *Post) error
		Delete(context.Context, int64) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]*PostWithMetadata, error)
	}

	Users interface {
		create(context.Context, *sql.Tx, *User) error
		createUserInvitation(context.Context, *sql.Tx, int64, string, time.Duration) error
		CreateAndInvite(context.Context, *User, string, time.Duration) error

		Get(context.Context, int64) (*User, error)
		ActivateAndClean(context.Context, string) error

		GetByUsername(context.Context, string) (*User, error)
		GetByEmail(context.Context, string) (*User, error)

		DeleteInvitation(context.Context, int64) error
		DeleteUser(context.Context, int64) error
		DeleteUserAndInvitation(context.Context, int64) error
	}

	Comments interface {
		Create(context.Context, *Comment) error
		GetByID(context.Context, int64, int64) (*Comment, error)
		GetByPostID(context.Context, int64) ([]CommentWithUser, error)
		DeleteByPostID(context.Context, int64) error
	}

	Followers interface {
		FollowUser(context.Context, int64, int64) error
		UnfollowUser(context.Context, int64, int64) error
		GetFollowers(context.Context, int64) ([]Follower, error)
	}

	Sessions interface {
		CreateSession(context.Context, *SessionData) error
		GetSessionByTokenHash(context.Context, *SessionData) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowersStore{db: db},
		Sessions:  &SessionsStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		return err
		// if rbErr := tx.Rollback(); rbErr != nil {
		// return rbErr
		// }
	}

	return tx.Commit()
}
