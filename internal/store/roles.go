package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// UserWithRole represents a user entity enriched with their full role permissions and details.
type UserWithRole struct {
	User

	Role Role
}

// Role defines the authorization role structure stored in the database.
type Role struct {
	Name        string `json:"name"`
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Level       int32  `json:"level"`
}

// RolesStore handles database operations for roles and permissions.
type RolesStore struct {
	db *sql.DB
}

// GetRoleByName retrieves a role entity by its unique name string.
func (s *RolesStore) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT id, description, level
		FROM roles 
		WHERE name = $1;
	`
	role := &Role{
		Name: name,
	}

	err := s.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Description, &role.Level)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get role by name %s: %w", name, err)
		}
	}

	return role, nil
}

// GetRoleByID retrieves a role entity by its unique ID.
func (s *RolesStore) GetRoleByID(ctx context.Context, id int64) (*Role, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		SELECT name, description, level
		FROM roles 
		WHERE id = $1;
	`
	role := &Role{
		ID: id,
	}

	err := s.db.QueryRowContext(ctx, query, id).Scan(&role.Name, &role.Description, &role.Level)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get role by id %d: %w", id, err)
		}
	}

	return role, nil
}
