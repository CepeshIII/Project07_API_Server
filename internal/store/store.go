package store

import (
	"context"
	"database/sql"
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
		Create(context.Context, *User) error
		Get(context.Context, int64) (*User, error)
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
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowersStore{db: db},
	}
}
