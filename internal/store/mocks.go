package store

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

func NewMockStorage() Storage {
	return Storage{
		Posts:     &MockPostStore{},
		Users:     &MockUserStore{},
		Comments:  &MockCommentStore{},
		Followers: &MockFollowersStore{},
		Sessions:  &MockSessionsStore{},
		Roles:     &MockRolesStore{},
	}
}

type MockPostStore struct{}

func (*MockPostStore) Create(context.Context, *Post) error {
	return nil
}

func (*MockPostStore) GetByID(context.Context, int64) (*Post, error) {
	return nil, nil
}

func (*MockPostStore) Update(context.Context, int64, *Post) error {
	return nil
}

func (*MockPostStore) Delete(context.Context, int64) error {
	return nil
}

func (*MockPostStore) GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]*PostWithMetadata, error) {
	return nil, nil
}

type MockCommentStore struct{}

func (*MockCommentStore) Create(context.Context, *Comment) error {
	return nil
}

func (*MockCommentStore) GetByID(context.Context, int64, int64) (*Comment, error) {
	return nil, nil
}

func (*MockCommentStore) GetByPostID(context.Context, int64) ([]CommentWithUser, error) {
	return nil, nil
}

func (*MockCommentStore) DeleteByPostID(context.Context, int64) error {
	return nil
}

type MockFollowersStore struct {
	mock.Mock
}

func (m *MockFollowersStore) FollowUser(ctx context.Context, userID, followerID int64) error {
	args := m.Called(ctx, userID, followerID)
	return args.Error(0)
}

func (m *MockFollowersStore) UnfollowUser(ctx context.Context, userID, followerID int64) error {
	args := m.Called(ctx, userID, followerID)
	return args.Error(0)
}

func (m *MockFollowersStore) GetFollowers(ctx context.Context, userID int64) ([]Follower, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Follower), args.Error(1)
}

type MockSessionsStore struct{}

func (*MockSessionsStore) CreateSession(context.Context, *SessionData) error {
	return nil
}

func (*MockSessionsStore) GetSessionByTokenHash(context.Context, *SessionData) error {
	return nil
}

type MockRolesStore struct {
	mock.Mock
}

func (s *MockRolesStore) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	args := s.Called(ctx, name)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Role), args.Error(1)
}

func (s *MockRolesStore) GetRoleByID(ctx context.Context, roleID int64) (*Role, error) {
	args := s.Called(ctx, roleID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Role), args.Error(1)
}

// MockUserStore — динамічний мок для Users
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, userWithRole *UserWithRole, token string, exp time.Duration) error {
	args := m.Called(ctx, userWithRole, token, exp)
	return args.Error(0)
}

func (m *MockUserStore) Get(ctx context.Context, userID int64) (*User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) ActivateAndClean(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) GetUserWithRole(ctx context.Context, userID int64) (*UserWithRole, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserWithRole), args.Error(1)
}

func (m *MockUserStore) DeleteInvitation(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserStore) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserStore) DeleteUserAndInvitation(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
