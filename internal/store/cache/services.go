package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// CheckConnection verifies that the Redis client can successfully connect, write, and read.
func CheckConnection(ctx context.Context, client *redis.Client) error {
	// Alternatively, you can use client.Ping(ctx).Err() for a standard lightweight check.
	testKey := "foo"
	testValue := "bar"

	if err := client.Set(ctx, testKey, testValue, 0).Err(); err != nil {
		return fmt.Errorf("failed to set test key in redis connection check: %w", err)
	}

	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get test key in redis connection check: %w", err)
	}

	if val != testValue {
		return fmt.Errorf("redis connection check failed (expected %q, got %q): %w", testValue, val, ErrRedisValueMismatch)
	}

	// Clean up the test key and handle the error/result properly
	if err := client.Del(ctx, testKey).Err(); err != nil {
		return fmt.Errorf("failed to cleanup test key in redis connection check: %w", err)
	}

	return nil
}
