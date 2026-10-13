package cache

import (
	"context"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/stretchr/testify/mock"
)

// NewMockStorage creates and returns a new Storage instance populated with mock stores.
func NewMockStorage() Storage {
	return Storage{
		Users: &MockUserStore{},
		Roles: &MockRolesStore{},
	}
}

// MockUserStore is a mock implementation of the user store for testing.
type MockUserStore struct {
	mock.Mock
}

// GetUserByID mocks fetching a user by their unique identifier.
func (s *MockUserStore) GetUserByID(ctx context.Context, userID int64) (*store.User, error) {
	args := s.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	user, ok := args.Get(0).(*store.User)
	if !ok {
		return nil, args.Error(1)
	}

	return user, args.Error(1)
}

// SetUser mocks saving or updating a user in the cache/store.
func (s *MockUserStore) SetUser(ctx context.Context, user *store.User) error {
	args := s.Called(ctx, user)

	return args.Error(0)
}

// MockRolesStore is a mock implementation of the roles store for testing.
type MockRolesStore struct {
	mock.Mock
}

// GetRoleByName mocks retrieving a role by its name.
func (s *MockRolesStore) GetRoleByName(ctx context.Context, name string) (*store.Role, error) {
	args := s.Called(ctx, name)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	role, ok := args.Get(0).(*store.Role)
	if !ok {
		return nil, args.Error(1)
	}

	return role, args.Error(1)
}

// GetRoleByID mocks retrieving a role by its unique identifier.
func (s *MockRolesStore) GetRoleByID(ctx context.Context, roleID int64) (*store.Role, error) {
	args := s.Called(ctx, roleID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	role, ok := args.Get(0).(*store.Role)
	if !ok {
		return nil, args.Error(1)
	}

	return role, args.Error(1)
}

// SetRole mocks saving or updating a role in the cache/store.
func (s *MockRolesStore) SetRole(ctx context.Context, role *store.Role) error {
	args := s.Called(ctx, role)

	return args.Error(0)
}
