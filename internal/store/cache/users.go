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

type UserStore struct {
	rdb *redis.Client
}

const UserExpTime = time.Minute
const UserCacheName = "user"

func (s UserStore) GetUserByID(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := userIDKey(userID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s UserStore) SetUser(ctx context.Context, user *store.User) error {
	cacheKey := userIDKey(user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return nil
	}

	return s.rdb.Set(ctx, cacheKey, json, UserExpTime).Err()
}

func userIDKey(userID int64) string {
	return fmt.Sprintf("user-id-%d", userID)
}
