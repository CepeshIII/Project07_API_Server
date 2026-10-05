package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/go-chi/chi/v5"
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
				decodeAndValidate(t, rr.Body.String(), &response)

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
				var errorEnvelope ErrorEnvelope
				decodeAndValidate(t, rr.Body.String(), &errorEnvelope)
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
		name                string
		expectedStatus      int
		mockSetup           func(m *store.MockUserStore)
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "successful activation of user (200)",
			expectedStatus: http.StatusOK,
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "not found activation of user (404)",
			expectedStatus: http.StatusNotFound,
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(store.ErrNotFound)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},

		{
			name:           "invalid activation of user (400)",
			expectedStatus: http.StatusBadRequest,
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(store.ErrorInvalidToken)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "internal server error (500)",
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func(m *store.MockUserStore) {
				expectedHash := string(HashToken("test-token"))
				m.On("ActivateAndClean", mock.Anything, expectedHash).Return(errors.New("db connection lost"))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
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

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestFollowUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}
	authUserID := int64(1)

	tests := []struct {
		name                string
		expectedStatus      int
		targetIDParam       string
		mockSetup           func(f *store.MockFollowersStore)
		contextSetup        func(r *http.Request) *http.Request
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success follow user",
			expectedStatus: http.StatusCreated,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Target user missing from context",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Auth userID missing from context",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "User cannot follow itself",
			expectedStatus: http.StatusBadRequest,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, targetUser.ID)) // same as target user ID
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Already following (Conflict)",
			expectedStatus: http.StatusConflict,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(store.ErrorConflict)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Internal server error on Follow user",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("FollowUser", mock.Anything, authUserID, targetUser.ID).Return(errors.New("database error"))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
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

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestUnfollowUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}
	authUserID := int64(1)

	tests := []struct {
		name                string
		expectedStatus      int
		targetIDParam       string
		mockSetup           func(f *store.MockFollowersStore)
		contextSetup        func(r *http.Request) *http.Request
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success unfollow user",
			expectedStatus: http.StatusCreated,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Target user missing from context",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Auth userID missing from context",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "User cannot unfollow itself",
			expectedStatus: http.StatusBadRequest,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, targetUser.ID)) // same as target user ID
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Already not following (Not Found)",
			expectedStatus: http.StatusNotFound,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(store.ErrNotFound)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Internal server error on Unfollow user",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  "2",
			contextSetup: func(r *http.Request) *http.Request {
				r = r.WithContext(context.WithValue(r.Context(), authUserIDCtxKey, authUserID))
				r = r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
				return r
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("UnfollowUser", mock.Anything, authUserID, targetUser.ID).Return(errors.New("database error"))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
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

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestGetFollowersHandler(t *testing.T) {
	targetUser := &store.User{ID: 1, Username: "targetUser"}

	tests := []struct {
		name                string
		expectedStatus      int
		contextSetup        func(r *http.Request) *http.Request
		mockSetup           func(f *store.MockFollowersStore)
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success get followers",
			expectedStatus: http.StatusOK,
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockFollowersStore) {
				followersList := []store.Follower{
					{UserSummary: &store.UserSummary{ID: 1, Username: "follower1"}},
					{UserSummary: &store.UserSummary{ID: 2, Username: "follower2"}},
					{UserSummary: &store.UserSummary{ID: 2, Username: "follower2"}},
				}
				f.On("GetFollowers", mock.Anything, targetUser.ID).Return(followersList, nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response PostsEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Target user missing from context",
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, nil))
			},
			mockSetup: func(f *store.MockFollowersStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response PostsEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Internal DB Error",
			expectedStatus: http.StatusInternalServerError,
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockFollowersStore) {
				f.On("GetFollowers", mock.Anything, targetUser.ID).Return(nil, errors.New("db error"))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response PostsEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
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

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	targetUser := &store.User{ID: 2, Username: "followed_user"}

	tests := []struct {
		name                string
		expectedStatus      int
		targetIDParam       string
		mockSetup           func(f *store.MockUserStore)
		contextSetup        func(r *http.Request) *http.Request
		dataValidationSetup func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:           "Success Delete user",
			expectedStatus: http.StatusOK,
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockUserStore) {
				f.On("DeleteUser", mock.Anything, targetUser.ID).Return(nil)
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Target user missing from context",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r
			},
			mockSetup: func(f *store.MockUserStore) {},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
		},
		{
			name:           "Internal server error when db error",
			expectedStatus: http.StatusInternalServerError,
			targetIDParam:  strconv.FormatInt(targetUser.ID, 10),
			contextSetup: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), targetUserCtxKey, targetUser))
			},
			mockSetup: func(f *store.MockUserStore) {
				f.On("DeleteUser", mock.Anything, targetUser.ID).Return(errors.New("database error"))
			},
			dataValidationSetup: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response MessageEnvelope
				decodeAndValidate(t, rr.Body.String(), &response)
			},
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
				r.Delete("/", app.deleteUserHandler)
			})

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.targetIDParam, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUsers.AssertExpectations(t)

			if tt.dataValidationSetup != nil {
				tt.dataValidationSetup(t, rr)
			}
		})
	}
}
