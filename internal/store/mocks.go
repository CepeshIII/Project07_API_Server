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

// --- MockPostStore ---

type MockPostStore struct {
	mock.Mock
}

func (m *MockPostStore) Create(ctx context.Context, post *PostModel) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostStore) GetPost(ctx context.Context, id int64) (*Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Post), args.Error(1)
}

func (m *MockPostStore) GetPostModel(ctx context.Context, id int64) (*PostModel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PostModel), args.Error(1)
}

func (m *MockPostStore) Update(ctx context.Context, id int64, req *UpdatePostRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockPostStore) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostStore) GetUserFeed(ctx context.Context, userID int64, query PaginatedFeedQuery) ([]*PostFeedItem, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*PostFeedItem), args.Error(1)
}

func (m *MockPostStore) GetAllPosts(ctx context.Context, query PaginatedFeedQuery) ([]*PostFeedItem, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*PostFeedItem), args.Error(1)
}

// --- MockCommentStore ---

type MockCommentStore struct {
	mock.Mock
}

func (m *MockCommentStore) Create(ctx context.Context, comment *Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentStore) GetByID(ctx context.Context, postID, commentID int64) (*Comment, error) {
	args := m.Called(ctx, postID, commentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Comment), args.Error(1)
}

func (m *MockCommentStore) GetByPostID(ctx context.Context, postID int64) ([]CommentWithUser, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CommentWithUser), args.Error(1)
}

func (m *MockCommentStore) DeleteByPostID(ctx context.Context, postID int64) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

// --- MockFollowersStore ---

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

func (m *MockFollowersStore) GetFollowerModels(ctx context.Context, userID int64) ([]FollowerModel, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]FollowerModel), args.Error(1)
}

// --- MockSessionsStore ---

type MockSessionsStore struct {
	mock.Mock
}

func (m *MockSessionsStore) CreateSession(ctx context.Context, data *SessionData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockSessionsStore) GetSessionByTokenHash(ctx context.Context, data *SessionData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// --- MockRolesStore ---

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

// --- MockUserStore ---

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
