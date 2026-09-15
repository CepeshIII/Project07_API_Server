package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/redis/go-redis/v9"
)

type RolesStore struct {
	rdb *redis.Client
}

func (s *RolesStore) GetRoleByName(ctx context.Context, roleName string) (*store.Role, error) {
	cacheKey := roleNameKey(roleName)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var role store.Role
	if data != "" {
		err := json.Unmarshal([]byte(data), &role)
		if err != nil {
			return nil, err
		}
	}

	return &role, nil
}

func (s *RolesStore) GetRoleByID(ctx context.Context, roleID int64) (*store.Role, error) {
	cacheKey := roleIDKey(roleID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	var role store.Role
	if data != "" {
		err := json.Unmarshal([]byte(data), &role)
		if err != nil {
			return nil, err
		}
	}

	return &role, nil
}

func (s *RolesStore) SetRole(ctx context.Context, role *store.Role) error {
	cacheByIDKey := roleIDKey(role.ID)
	cacheByNameKey := roleNameKey(role.Name)

	json, err := json.Marshal(role)
	if err != nil {
		return err
	}

	if err := s.rdb.Set(ctx, cacheByIDKey, json, UserExpTime).Err(); err != nil {
		return err

	}

	if err := s.rdb.Set(ctx, cacheByNameKey, json, UserExpTime).Err(); err != nil {
		return err
	}

	return nil
}

func roleIDKey(roleID int64) string {
	return fmt.Sprintf("role-id-%d", roleID)
}

func roleNameKey(roleName string) string {
	return fmt.Sprintf("role-name-%s", roleName)
}
