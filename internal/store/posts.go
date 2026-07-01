package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("record not found")

type PostWithComments struct {
	Post     *Post             `json:"post_data"`
	Comments []CommentWithUser `json:"post_comments"`
}

type Post struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	Title   string `json:"title"`
	UserID  int64  `json:"user_id"`

	Tags []string `json:"tags"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
	INSERT INTO posts (content, title, user_id, tags)
	VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	return err
}

func (s *PostStore) GetByID(ctx context.Context, postID int64) (*Post, error) {
	post := Post{}

	query := `
	SELECT id, content, title, user_id, tags, created_at, updated_at
	FROM posts 
	WHERE id = $1
	`

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,

		pq.Array(&post.Tags),

		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, postID int64, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, tags = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, title, content, tags, user_id, created_at, updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		postID,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	default:
		return err
	}
}

func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
	`
	res, err := s.db.ExecContext(ctx, query, postID)

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
