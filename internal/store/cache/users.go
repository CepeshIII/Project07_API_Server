package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/redis/go-redis/v9"
)

// UserExpTime defines the expiration duration for cached user entities.
const UserExpTime = time.Minute

// UserStore handles Redis caching operations for user entities.
type UserStore struct {
	rdb *redis.Client
}

// GetUserByID retrieves a cached user entity by their unique ID.
func (s *UserStore) GetUserByID(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := userIDKey(userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMissed
		}

		return nil, fmt.Errorf("failed to get user by id %d from cache: %w", userID, err)
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user by id %d from cache: %w", userID, err)
		}
	}

	return &user, nil
}

// SetUser stores a user entity in the cache with the defined expiration time.
func (s *UserStore) SetUser(ctx context.Context, user *store.User) error {
	cacheKey := userIDKey(user.ID)

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user %d for cache: %w", user.ID, err)
	}

	if err := s.rdb.Set(ctx, cacheKey, jsonBytes, UserExpTime).Err(); err != nil {
		return fmt.Errorf("failed to set user by id %d in cache: %w", user.ID, err)
	}

	return nil
}

func userIDKey(userID int64) string {
	return fmt.Sprintf("user-id-%d", userID)
}
