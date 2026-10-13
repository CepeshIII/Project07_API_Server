package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// FollowersStore handles database operations for user follow relationships.
type FollowersStore struct {
	db *sql.DB
}

// FollowerModel represents the raw database structure of a follower relationship.
type FollowerModel struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

// Follower represents a follower enriched with summary user details.
type Follower struct {
	UserSummary *UserSummary `json:"user_summary"`
	CreatedAt   string       `json:"created_at"`
}

// FollowUser creates a relationship where followerID follows followeeID (userID).
func (s *FollowersStore) FollowUser(ctx context.Context, followerID, followeeID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		followeeID,
		followerID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrResourceConflict
		}

		return fmt.Errorf("failed to follow user %d by follower %d: %w", followeeID, followerID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for follow user %d: %w", followeeID, err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// UnfollowUser removes a follow relationship where followerID stops following followeeID.
func (s *FollowersStore) UnfollowUser(ctx context.Context, followerID, followeeID int64) error {
	query := `
	DELETE FROM followers
		WHERE user_id = $1 and follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(
		ctx,
		query,
		followeeID,
		followerID,
	)
	if err != nil {
		return fmt.Errorf("failed to unfollow user %d by follower %d: %w", followeeID, followerID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for unfollow user %d: %w", followeeID, err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetFollowers retrieves a list of followers with their user summaries for a given user ID.
func (s *FollowersStore) GetFollowers(ctx context.Context, userID int64) ([]Follower, error) {
	query := `
	SELECT f.follower_id, f.created_at, u.username
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers for user %d: %w", userID, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var followers []Follower

	for rows.Next() {
		follower := Follower{
			UserSummary: &UserSummary{},
		}

		err := rows.Scan(&follower.UserSummary.ID, &follower.CreatedAt, &follower.UserSummary.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower row for user %d: %w", userID, err)
		}

		followers = append(followers, follower)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error getting followers for user %d: %w", userID, err)
	}

	return followers, nil
}

// GetFollowerModels retrieves raw follower model entries for a given user ID.
func (s *FollowersStore) GetFollowerModels(ctx context.Context, userID int64) ([]FollowerModel, error) {
	query := `
    SELECT user_id, follower_id, created_at FROM followers
        WHERE user_id = $1
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query follower models for user %d: %w", userID, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var followers []FollowerModel

	for rows.Next() {
		var follower FollowerModel

		err := rows.Scan(&follower.UserID, &follower.FollowerID, &follower.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follower model row for user %d: %w", userID, err)
		}

		followers = append(followers, follower)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error getting follower models for user %d: %w", userID, err)
	}

	return followers, nil
}
