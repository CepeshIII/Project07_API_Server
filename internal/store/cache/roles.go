package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/redis/go-redis/v9"
)

// RolesStore handles Redis caching operations for role entities.
type RolesStore struct {
	rdb *redis.Client
}

// GetRoleByName retrieves a cached role by its name.
func (s *RolesStore) GetRoleByName(ctx context.Context, roleName string) (*store.Role, error) {
	cacheKey := roleNameKey(roleName)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMissed
		}

		return nil, fmt.Errorf("failed to get role by name %s from cache: %w", roleName, err)
	}

	var role store.Role
	if data != "" {
		if err := json.Unmarshal([]byte(data), &role); err != nil {
			return nil, fmt.Errorf("failed to unmarshal role by name %s from cache: %w", roleName, err)
		}
	}

	return &role, nil
}

// GetRoleByID retrieves a cached role by its unique ID.
func (s *RolesStore) GetRoleByID(ctx context.Context, roleID int64) (*store.Role, error) {
	cacheKey := roleIDKey(roleID)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMissed
		}

		return nil, fmt.Errorf("failed to get role by id %d from cache: %w", roleID, err)
	}

	var role store.Role
	if data != "" {
		if err := json.Unmarshal([]byte(data), &role); err != nil {
			return nil, fmt.Errorf("failed to unmarshal role by id %d from cache: %w", roleID, err)
		}
	}

	return &role, nil
}

// SetRole stores a role entity in cache indexed by both ID and name keys.
func (s *RolesStore) SetRole(ctx context.Context, role *store.Role) error {
	cacheByIDKey := roleIDKey(role.ID)
	cacheByNameKey := roleNameKey(role.Name)

	jsonBytes, err := json.Marshal(role)
	if err != nil {
		return fmt.Errorf("failed to marshal role %d for cache: %w", role.ID, err)
	}

	if err := s.rdb.Set(ctx, cacheByIDKey, jsonBytes, UserExpTime).Err(); err != nil {
		return fmt.Errorf("failed to set role by id %d in cache: %w", role.ID, err)
	}

	if err := s.rdb.Set(ctx, cacheByNameKey, jsonBytes, UserExpTime).Err(); err != nil {
		return fmt.Errorf("failed to set role by name %s in cache: %w", role.Name, err)
	}

	return nil
}

func roleIDKey(roleID int64) string {
	return fmt.Sprintf("role-id-%d", roleID)
}

func roleNameKey(roleName string) string {
	return "role-name-" + roleName
}
