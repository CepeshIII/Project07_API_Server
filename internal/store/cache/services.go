package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func CheckConnection(ctx context.Context, client *redis.Client) error {
	testKey := "foo"
	testValue := "bar"

	err := client.Set(ctx, testKey, testValue, 0).Err()
	if err != nil {
		return err
	}

	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		return err
	}

	if val != testValue {
		return fmt.Errorf("redis connection check failed: value mismatch (expected %q, got %q)", testValue, val)
	}

	// Clean up the test key (passing testKey as an argument)
	client.Del(ctx, testKey)

	return nil
}
