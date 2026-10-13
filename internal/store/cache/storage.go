// Package cache provides caching storage implementations, Redis clients,
// and mock stores for testing purposes.
package cache

import (
	"context"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/redis/go-redis/v9"
)

// Storage aggregates all Redis caching stores for users and roles.
type Storage struct {
	Users interface {
		GetUserByID(ctx context.Context, userID int64) (*store.User, error)
		SetUser(ctx context.Context, user *store.User) error
	}

	Roles interface {
		GetRoleByName(ctx context.Context, roleName string) (*store.Role, error)
		GetRoleByID(ctx context.Context, roleID int64) (*store.Role, error)
		SetRole(ctx context.Context, role *store.Role) error
	}
}

// NewRedisStore initializes and returns a new Storage aggregation instance backed by Redis.
func NewRedisStore(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
		Roles: &RolesStore{rdb: rdb},
	}
}
