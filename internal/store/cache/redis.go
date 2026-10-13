package cache

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrRedisValueMismatch is returned when the retrieved value from Redis does not match expectations.
	ErrRedisValueMismatch = errors.New("redis value mismatch")

	// ErrCacheMissed is returned when a requested key is not found in the cache.
	ErrCacheMissed = errors.New("cache miss")
)

// NewRedisClient initializes and returns a new configured Redis client instance.
func NewRedisClient(addr, pw string, db int) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: pw,
			DB:       db,
		})
}
