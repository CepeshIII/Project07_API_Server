package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/testify/v2/require"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockServer struct {
	mock.Mock
}

func TestGetUserHandler(t *testing.T) {
	user := &store.User{
		ID:       1,
		Username: "user",
	}

	tests := []struct {
		name                string
		userIDParam         string
		expectedStatus      int
		contextSetup        func(*http.Request) *http.Request
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "successful retrieval of user",
			userIDParam:    strconv.Itoa(int(user.ID)),
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, user))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response UserEnvelope
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, user.ID, response.Data.ID)
				assert.Equal(t, user.Username, response.Data.Username)
			},
		},
		{
			name:           "target user missing from context",
			userIDParam:    strconv.Itoa(int(user.ID)),
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				return r // context without target user
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "the server encountered a problem")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				logger: zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tt.contextSetup != nil {
						r = tt.contextSetup(r)
					}
					next.ServeHTTP(w, r)
				})
			})
			r.Get("/users/{userID}", app.getUserHandler)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", tt.userIDParam), nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestActivateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(m *store.MockUserStore)
		expectedStatus int
	}{
		{
			name: "successful activation of user (200)",
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found activation of user (404)",
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(store.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},

		{
			name: "invalid activation of user (400)",
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(store.ErrorInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal server error (500)",
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(errors.New("db connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsers := new(store.MockUserStore)
			tt.mockSetup(mockUsers)

			app := &application{
				logger: zap.NewNop().Sugar(),
				store: store.Storage{
					Users: mockUsers,
				},
			}

			r := chi.NewRouter()
			r.Put("/users/activate/{token}", app.activateUserHandler)

			req := httptest.NewRequest(http.MethodPut, "/users/activate/test-token", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUsers.AssertExpectations(t)
		})
	}
}

func TestFollowUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}
	authUserID := int64(1)

	tests := []struct {
		name           string
		targetIDParam  string
		mockSetup      func(f *store.MockFollowersStore)
		contextSetup   func(r *http.Request) *http.Request
		expectedStatus int
	}{
		{
			name:          "Success follow user",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:          "Target user missing from context",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "Auth userID missing from context",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "User cannot follow itself",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, targetUser.ID)) // same as target user ID
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Already following (Conflict)",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(store.ErrorConflict)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:          "Internal server error on Follow user",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFollowers := new(store.MockFollowersStore)

			tt.mockSetup(mockFollowers)

			app := &application{
				store: store.Storage{
					Followers: mockFollowers,
				},
				logger: zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						req = tt.contextSetup(req)
						next.ServeHTTP(w, req)
					})
				})
				r.Put("/follow", app.followUserHandler)
			})

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.targetIDParam+"/follow", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockFollowers.AssertExpectations(t)
		})
	}
}

func TestUnfollowUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}
	authUserID := int64(1)

	tests := []struct {
		name           string
		targetIDParam  string
		mockSetup      func(f *store.MockFollowersStore)
		contextSetup   func(r *http.Request) *http.Request
		expectedStatus int
	}{
		{
			name:          "Success unfollow user",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:          "Target user missing from context",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "Auth userID missing from context",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "User cannot unfollow itself",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, targetUser.ID)) // same as target user ID
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Already not following (Not Found)",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(store.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "Internal server error on Unfollow user",
			targetIDParam: "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFollowers := new(store.MockFollowersStore)

			tt.mockSetup(mockFollowers)

			app := &application{
				store: store.Storage{
					Followers: mockFollowers,
				},
				logger: zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						req = tt.contextSetup(req)
						next.ServeHTTP(w, req)
					})
				})
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.targetIDParam+"/unfollow", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockFollowers.AssertExpectations(t)
		})
	}
}

func TestGetFollowersHandler(t *testing.T) {
	targetUser := &store.User{ID: 1, Username: "targetUser"}

	tests := []struct {
		name           string
		contextSetup   func(r *http.Request) *http.Request
		mockSetup      func(f *store.MockFollowersStore)
		expectedStatus int
	}{
		{
			name: "Success get followers",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockFollowersStore) {
				followersList := []store.Follower{
					{FollowerID: 2, UserID: targetUser.ID},
					{FollowerID: 3, UserID: targetUser.ID},
				}
				f.On("GetFollowers", mock.Anything, targetUser.ID).Return(followersList, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Target user missing from context",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, nil))
			},
			mockSetup:      func(f *store.MockFollowersStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Internal DB Error",
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("GetFollowers", mock.Anything, targetUser.ID).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFollowers := new(store.MockFollowersStore)
			tt.mockSetup(mockFollowers)

			app := &application{
				store:  store.Storage{Followers: mockFollowers},
				logger: zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						req = tt.contextSetup(req)
						next.ServeHTTP(w, req)
					})
				})
				r.Get("/followers", app.getFollowersHandler)
			})

			req := httptest.NewRequest(http.MethodGet, "/users/1/followers", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockFollowers.AssertExpectations(t)
		})
	}
}

func TestCheckUserOwnershipMiddleware(t *testing.T) {
	authUser := &store.User{ID: 1, Username: "auth_user", RoleID: 1}
	adminUser := &store.User{ID: 1, Username: "admin", RoleID: 3}
	targetUser := &store.User{ID: 2, Username: "target_user", RoleID: 1}

	userRole := &store.Role{ID: 1, Name: "user", Level: 1}
	adminRole := &store.Role{ID: 3, Name: "admin", Level: 3}

	tests := []struct {
		name           string
		targetIDParam  string
		cachingEnabled bool
		expectedStatus int
		contextSetup   func(r *http.Request) *http.Request
		mockSetup      func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore)
	}{
		{
			name:           "should allow access when user is owner (cache disabled)",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, authUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				su.On("Get", mock.Anything, authUser.ID).Return(authUser, nil)
				sr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, nil)
			},
		},
		{
			name:           "should allow access when user is owner (cache enabled)",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: true,
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, authUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				cu.On("GetUserByID", mock.Anything, authUser.ID).Return(authUser, nil)
				cr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, nil)
			},
		},
		{
			name:           "should deny access when user is not owner and lacks permission (cache disabled)",
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusForbidden,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, targetUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				su.On("Get", mock.Anything, authUser.ID).Return(authUser, nil)
				sr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, nil)
				sr.On("GetRoleByName", mock.Anything, adminRole.Name).Return(adminRole, nil)
			},
		},
		{
			name:           "should deny access when user is not owner and lacks permission (cache enabled)",
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			cachingEnabled: true,
			expectedStatus: http.StatusForbidden,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, targetUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				cu.On("GetUserByID", mock.Anything, authUser.ID).Return(authUser, nil)
				cr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, nil)
				cr.On("GetRoleByName", mock.Anything, adminRole.Name).Return(adminRole, nil)
			},
		},
		{
			name:           "should allow access when user is not owner but has required permission (cache disabled)",
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, adminUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, targetUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				su.On("Get", mock.Anything, adminUser.ID).Return(adminUser, nil)
				sr.On("GetRoleByID", mock.Anything, adminUser.RoleID).Return(adminRole, nil)
				sr.On("GetRoleByName", mock.Anything, adminRole.Name).Return(adminRole, nil)
			},
		},
		{
			name:           "should allow access when user is not owner but has required permission (cache enabled)",
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			cachingEnabled: true,
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, adminUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, targetUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				cu.On("GetUserByID", mock.Anything, adminUser.ID).Return(adminUser, nil)
				cr.On("GetRoleByID", mock.Anything, adminUser.RoleID).Return(adminRole, nil)
				cr.On("GetRoleByName", mock.Anything, adminRole.Name).Return(adminRole, nil)
			},
		},
		{
			name:           "should return 500 when target user is missing from context",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
			},
		},
		{
			name:           "should return 500 when auth user ID is missing from context",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), targetUserCtxKey, authUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
			},
		},
		{
			name:           "should return 500 when db error(cache disable)",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: false,
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, authUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				su.On("Get", mock.Anything, authUser.ID).Return(authUser, errors.New("db error"))
			},
		},
		{
			name:           "should work if cache has errror but db works",
			targetIDParam:  strconv.FormatInt(authUser.ID, 10),
			cachingEnabled: true,
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), authUserIDCtxKey, authUser.ID)
				ctx = context.WithValue(ctx, targetUserCtxKey, authUser)
				return r.WithContext(ctx)
			},
			mockSetup: func(su *store.MockUserStore, cu *cache.MockUserStore, sr *store.MockRolesStore, cr *cache.MockRolesStore) {
				su.On("Get", mock.Anything, authUser.ID).Return(authUser, nil)
				sr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, nil)

				cu.On("GetUserByID", mock.Anything, authUser.ID).Return(authUser, errors.New("cache error"))
				cr.On("GetRoleByID", mock.Anything, userRole.ID).Return(userRole, errors.New("cache error"))

				cu.On("SetUser", mock.Anything, authUser).Return(errors.New("cache error"))
				cr.On("SetRole", mock.Anything, userRole).Return(errors.New("cache error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserStore := new(store.MockUserStore)
			mockCacheUserStore := new(cache.MockUserStore)
			mockRolesStore := new(store.MockRolesStore)
			mockCacheRoleStore := new(cache.MockRolesStore)

			tt.mockSetup(mockUserStore, mockCacheUserStore, mockRolesStore, mockCacheRoleStore)

			app := &application{
				store: store.Storage{
					Users: mockUserStore,
					Roles: mockRolesStore,
				},
				cacheStorage: cache.Storage{
					Users: mockCacheUserStore,
					Roles: mockCacheRoleStore,
				},
				logger: zap.NewNop().Sugar(),
				config: config{
					redisCfg: redisConfig{
						enabled: tt.cachingEnabled,
					},
				},
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = app.jsonResponse(w, http.StatusOK, "OK")
			})

			r := chi.NewRouter()
			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						req = tt.contextSetup(req)
						next.ServeHTTP(w, req)
					})
				})
				r.Put("/test", app.checkUserOwnership("admin", nextHandler))
			})

			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.targetIDParam+"/test", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUserStore.AssertExpectations(t)
			mockCacheUserStore.AssertExpectations(t)
			mockRolesStore.AssertExpectations(t)
			mockCacheRoleStore.AssertExpectations(t)
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}

	tests := []struct {
		name           string
		targetIDParam  string
		mockSetup      func(f *store.MockUserStore)
		contextSetup   func(r *http.Request) *http.Request
		expectedStatus int
	}{
		{
			name:          "Success Delete user",
			targetIDParam: strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockUserStore) {
				f.On("DeleteUser", mock.Anything, targetUser.ID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Target user missing from context",
			targetIDParam: strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r
			},
			mockSetup:      func(f *store.MockUserStore) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "Internal server error when db error",
			targetIDParam: strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockUserStore) {
				f.On("DeleteUser", mock.Anything, targetUser.ID).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsers := new(store.MockUserStore)
			tt.mockSetup(mockUsers)

			app := &application{
				store: store.Storage{
					Users: mockUsers,
				},
				logger: zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						req = tt.contextSetup(req)
						next.ServeHTTP(w, req)
					})
				})
				r.Delete("/", app.deleteUserHandler) // Використовуємо Delete замість Put("/delete")
			})

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.targetIDParam, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUsers.AssertExpectations(t)
		})
	}
}
