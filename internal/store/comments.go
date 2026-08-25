package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type CommentWithUser struct {
	Comment
	Username string `json:"username"`
}

type CommentStore struct {
	db *sql.DB
}

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

	return err
}

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
		return nil, err
	}

	defer rows.Close()

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
			return nil, err
		}

		comments = append(comments, commentWithUser)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(comments) == 0 {
		return comments, ErrNotFound
	}

	return comments, nil
}

func (s *CommentStore) GetByID(ctx context.Context, postID int64, userID int64) (*Comment, error) {
	comment := Comment{}

	query := `
	SELECT id, post_id, user_id, content, created_at
	FROM comments 
	WHERE postID = $1 and userID = $2
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
			return nil, err
		}
	}

	return &comment, nil
}

func (s *CommentStore) DeleteByPostID(ctx context.Context, postID int64) error {
	query := `
		DELETE FROM comments
		WHERE post_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, postID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	default:
		return err
	}
}
