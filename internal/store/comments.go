package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Comment represents a comment entity on a post.
type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// CommentWithUser represents a comment enriched with the author's username.
type CommentWithUser struct {
	Comment

	Username string `json:"username"`
}

// CommentStore handles database operations for post comments.
type CommentStore struct {
	db *sql.DB
}

// Create inserts a new comment record into the database.
func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content)
		VALUES ($1, $2, $3) 
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment for post %d by user %d: %w", comment.PostID, comment.UserID, err)
	}

	return nil
}

// GetByPostID retrieves all comments associated with a specific post ID, ordered by creation date descending.
func (s *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]CommentWithUser, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username
		FROM comments c
		INNER JOIN users ON c.user_id = users.id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments for post %d: %w", postID, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	comments := []CommentWithUser{}

	for rows.Next() {
		var commentWithUser CommentWithUser

		err = rows.Scan(
			&commentWithUser.ID,
			&commentWithUser.PostID,
			&commentWithUser.UserID,
			&commentWithUser.Content,
			&commentWithUser.CreatedAt,
			&commentWithUser.Username,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row for post %d: %w", postID, err)
		}

		comments = append(comments, commentWithUser)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("row iteration error getting comments for post %d: %w", postID, err)
	}

	if len(comments) == 0 {
		return comments, ErrNotFound
	}

	return comments, nil
}

// GetByID retrieves a specific comment by post ID and user ID.
func (s *CommentStore) GetByID(ctx context.Context, postID int64, userID int64) (*Comment, error) {
	comment := Comment{}

	query := `
	SELECT id, post_id, user_id, content, created_at
	FROM comments 
	WHERE post_id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, postID, userID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get comment by post %d and user %d: %w", postID, userID, err)
		}
	}

	return &comment, nil
}

// DeleteByPostID removes all comments associated with a specific post ID.
func (s *CommentStore) DeleteByPostID(ctx context.Context, postID int64) error {
	query := `
		DELETE FROM comments
		WHERE post_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return fmt.Errorf("failed to delete comments for post %d: %w", postID, err)
		}
	}

	return nil
}
