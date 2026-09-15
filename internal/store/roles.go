package store

import (
	"context"
	"database/sql"
	"errors"
)

type UserWithRole struct {
	User
	Role Role
}

type Role struct {
	Name        string `json:"name"`
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Level       int32  `json:"level"`
}

type RolesStore struct {
	db *sql.DB
}

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
			return nil, err
		}
	}

	return role, nil
}

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
			return nil, err
		}
	}

	return role, nil
}
