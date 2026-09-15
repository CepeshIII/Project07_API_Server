package cache

import (
	"context"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Users interface {
		GetUserByID(context.Context, int64) (*store.User, error)
		SetUser(ctx context.Context, user *store.User) error
	}

	Roles interface {
		GetRoleByName(context.Context, string) (*store.Role, error)
		GetRoleByID(context.Context, int64) (*store.Role, error)
		SetRole(context.Context, *store.Role) error
	}
}

func NewRedisStore(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
		Roles: &RolesStore{rdb: rdb},
	}
}
