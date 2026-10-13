package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// PostModel represents the database entity and matches the exact columns of the posts table.
type PostModel struct {
	ID      int64  `db:"id"`
	Title   string `db:"title"`
	UserID  int64  `db:"user_id"`
	Content string `db:"content"`

	Tags []string `db:"tags"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Version int `db:"version"`
}

// Post represents Domain Model.
type Post struct {
	ID      int64  `example:"1"    json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Version int    `json:"version"`

	Tags []string `json:"tags"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	User *User `json:"user"`
}

// PostWithComments represents a post along with its associated comments.
type PostWithComments struct {
	Post     *Post             `json:"post_data"`
	Comments []CommentWithUser `json:"post_comments"`
}

// PostFeedItem represents a summary item for a post displayed within feeds.
type PostFeedItem struct {
	ID            int64        `json:"id"`
	Title         string       `json:"title"`
	Content       string       `json:"content"`
	Tags          []string     `json:"tags"`
	CreatedAt     string       `json:"created_at"`
	Version       int          `json:"version"`
	UserSummary   *UserSummary `json:"user_summary"`
	CommentsCount int          `json:"comments_count"`
}

// UpdatePostRequest defines the payload allowed when updating a post.
type UpdatePostRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`

	// Pro-tip: You can include Version here if you implement Optimistic Locking
	// to prevent race conditions when two people edit the post at the same time.
	Version int `json:"version"`
}

// PostStore handles database operations for posts.
type PostStore struct {
	db *sql.DB
}

// Create inserts a new post record into the database.
func (s *PostStore) Create(ctx context.Context, post *PostModel) error {
	query := `
    INSERT INTO posts (content, title, user_id, tags)
    VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

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
	if err != nil {
		return fmt.Errorf("failed to create post for user %d: %w", post.UserID, err)
	}

	return nil
}

// GetPost retrieves a full post entity by its unique ID, including author details.
func (s *PostStore) GetPost(ctx context.Context, postID int64) (*Post, error) {
	post := Post{
		User: &User{},
	}

	query := `
    SELECT p.id, p.content, p.title, p.tags, p.created_at, p.updated_at, p.version, u.id, u.username, u.email, u.created_at, u.is_active, u.password, u.role_id
    FROM posts p
    LEFT JOIN users u ON p.user_id = u.id
    WHERE p.id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
		&post.User.ID,
		&post.User.Username,
		&post.User.Email,
		&post.User.CreatedAt,
		&post.User.IsActive,
		&post.User.Password.hash,
		&post.User.RoleID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get post by id %d: %w", postID, err)
		}
	}

	return &post, nil
}

// GetPostModel retrieves the raw database model of a post by its unique ID.
func (s *PostStore) GetPostModel(ctx context.Context, postID int64) (*PostModel, error) {
	postModel := PostModel{}

	query := `
    SELECT id, content, title, user_id, tags, created_at, updated_at, version
    FROM posts 
    WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, postID).Scan(
		&postModel.ID,
		&postModel.Content,
		&postModel.Title,
		&postModel.UserID,
		pq.Array(&postModel.Tags),
		&postModel.CreatedAt,
		&postModel.UpdatedAt,
		&postModel.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get post model by id %d: %w", postID, err)
		}
	}

	return &postModel, nil
}

// Update modifies an existing post record using optimistic locking.
func (s *PostStore) Update(ctx context.Context, postID int64, post *UpdatePostRequest) error {
	query := `
        UPDATE posts
        SET title = $1, content = $2, tags = $3, updated_at = NOW(), version = version + 1
        WHERE id = $4 AND version = $5
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		postID,
		post.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to execute update for post %d: %w", postID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for post update %d: %w", postID, err)
	}

	if count == 0 {
		return ErrVersionConflict
	}

	return nil
}

// Delete removes a post record by its unique ID.
func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	query := `
        DELETE FROM posts
        WHERE id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("failed to execute delete for post %d: %w", postID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for post deletion %d: %w", postID, err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetUserFeed retrieves a paginated feed of posts for a given user based on followers and filters.
func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]*PostFeedItem, error) {
	sortOrder := SortModeDESC
	if fq.Sort == SortModeASC {
		sortOrder = SortModeASC
	}

	query := `
        SELECT 
            p.id, p.title, p.content, p.created_at, p.version, p.tags, u.username,
            COUNT(c.id) AS comments_count
        FROM posts p
        LEFT JOIN comments c ON c.post_id = p.id
        LEFT JOIN users u ON p.user_id = u.id
        JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
        WHERE 
            (COALESCE(NULLIF($6, ''), '') = '' OR p.created_at >= $6::timestamptz)
        AND
            (COALESCE(NULLIF($7, ''), '') = '' OR p.created_at <= $7::timestamptz)
        AND
            (p.tags = '{}' OR p.tags @> $4)
        AND 
            (p.content ILIKE '%' || $5 || '%' OR p.title ILIKE '%' || $5 || '%')
        AND 
            (f.user_id = $1 OR p.user_id = $1)
        GROUP BY p.id, u.username
        ORDER BY p.created_at ` + sortOrder + `
        LIMIT $2
        OFFSET $3
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		fq.Limit,
		fq.Offset,
		pq.Array(fq.Tags),
		fq.Query,
		fq.Since,
		fq.Until,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query user feed for user %d: %w", userID, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	feed := []*PostFeedItem{}

	for rows.Next() {
		postFeedItem := PostFeedItem{
			UserSummary: &UserSummary{},
		}

		err := rows.Scan(
			&postFeedItem.ID, &postFeedItem.Title, &postFeedItem.Content, &postFeedItem.CreatedAt, &postFeedItem.Version,
			pq.Array(&postFeedItem.Tags), &postFeedItem.UserSummary.Username, &postFeedItem.CommentsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post feed item for user %d: %w", userID, err)
		}

		feed = append(feed, &postFeedItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error for user feed %d: %w", userID, err)
	}

	return feed, nil
}

// GetAllPosts retrieves a paginated list of all posts across the system with optional filters.
func (s *PostStore) GetAllPosts(ctx context.Context, fq PaginatedFeedQuery) ([]*PostFeedItem, error) {
	sortOrder := SortModeDESC
	if fq.Sort == SortModeASC {
		sortOrder = SortModeASC
	}

	query := `
        SELECT 
            p.id, p.title, p.content, p.created_at, p.version, p.tags, u.username, p.user_id,
            COUNT(c.id) AS comments_count
        FROM posts p
        LEFT JOIN comments c ON c.post_id = p.id
        LEFT JOIN users u ON p.user_id = u.id
        WHERE 
            (COALESCE(NULLIF($5, ''), '') = '' OR p.created_at >= $5::timestamptz)
        AND
            (COALESCE(NULLIF($6, ''), '') = '' OR p.created_at <= $6::timestamptz)
        AND
            (p.tags = '{}' OR p.tags @> $3)
        AND 
            (p.content ILIKE '%' || $4 || '%' OR p.title ILIKE '%' || $4 || '%')
        GROUP BY p.id, u.username
        ORDER BY p.created_at ` + sortOrder + `
        LIMIT $1
        OFFSET $2
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		fq.Limit,
		fq.Offset,
		pq.Array(fq.Tags),
		fq.Query,
		fq.Since,
		fq.Until,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query all posts: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	feed := []*PostFeedItem{}

	for rows.Next() {
		postFeedItem := PostFeedItem{
			UserSummary: &UserSummary{},
		}

		err := rows.Scan(
			&postFeedItem.ID, &postFeedItem.Title, &postFeedItem.Content, &postFeedItem.CreatedAt, &postFeedItem.Version,
			pq.Array(&postFeedItem.Tags), &postFeedItem.UserSummary.Username, &postFeedItem.UserSummary.ID, &postFeedItem.CommentsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post feed item in all posts: %w", err)
		}

		feed = append(feed, &postFeedItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error in all posts: %w", err)
	}

	return feed, nil
}
