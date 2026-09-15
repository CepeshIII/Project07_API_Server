package cache

import (
	"context"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockStorage() Storage {
	return Storage{
		Users: &MockUserStore{},
		Roles: &MockRolesStore{},
	}
}

type MockUserStore struct {
	mock.Mock
}

func (s *MockUserStore) GetUserByID(ctx context.Context, userID int64) (*store.User, error) {
	args := s.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.User), args.Error(1)
}

func (s *MockUserStore) SetUser(ctx context.Context, user *store.User) error {
	args := s.Called(ctx, user)
	return args.Error(0)
}

type MockRolesStore struct {
	mock.Mock
}

func (s *MockRolesStore) GetRoleByName(ctx context.Context, name string) (*store.Role, error) {
	args := s.Called(ctx, name)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.Role), args.Error(1)
}

func (s *MockRolesStore) GetRoleByID(ctx context.Context, roleID int64) (*store.Role, error) {
	args := s.Called(ctx, roleID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*store.Role), args.Error(1)
}

func (s *MockRolesStore) SetRole(ctx context.Context, role *store.Role) error {
	args := s.Called(ctx, role)
	return args.Error(0)
}
