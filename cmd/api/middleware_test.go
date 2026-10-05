package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
	"github.com/CepeshIII/Project07_API_Server/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestRateLimiterMiddleWare(t *testing.T) {
	ip := "192.168.1.1"
	retryAfter := 10 * time.Second

	tests := []struct {
		Name           string
		ExpectedStatus int
		MockSetup      func(m *ratelimiter.MockRateLimiter)
		ResultCheck    func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			Name:           "should allow request when limiter allows",
			ExpectedStatus: http.StatusOK,
			MockSetup: func(m *ratelimiter.MockRateLimiter) {
				m.On("Allow", ip).Return(true, retryAfter)
			},
			ResultCheck: func(t *testing.T, rr *httptest.ResponseRecorder) {},
		},
		{
			Name:           "should reject request when limiter does not allow",
			ExpectedStatus: http.StatusTooManyRequests,
			MockSetup: func(m *ratelimiter.MockRateLimiter) {
				m.On("Allow", ip).Return(false, retryAfter)
			},
			ResultCheck: func(t *testing.T, rr *httptest.ResponseRecorder) {
				retryAfterHeader := rr.Header().Get("Retry-After")
				assert.NotEmpty(t, retryAfterHeader)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLimiter := ratelimiter.NewMockRateLimiter()
			test.MockSetup(mockLimiter)

			app := &application{
				rateLimiter: mockLimiter,
				logger:      zap.NewNop().Sugar(),
			}

			r := chi.NewRouter()
			r.With(app.rateLimiterMiddleWare).Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = ip
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, test.ExpectedStatus, rr.Code)
			mockLimiter.AssertExpectations(t)

			if test.ResultCheck != nil {
				test.ResultCheck(t, rr)
			}
		})
	}
}

func TestUserContextMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		userIDParam    string
		mockSetup      func(m *store.MockUserStore)
		expectedStatus int
	}{
		{
			name:        "successful retrieval of user (200)",
			userIDParam: "1",
			mockSetup: func(m *store.MockUserStore) {
				m.On("Get", mock.Anything, int64(1)).Return(&store.User{ID: 1, Username: "jack"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "user not found (404)",
			userIDParam: "999",
			mockSetup: func(m *store.MockUserStore) {
				m.On("Get", mock.Anything, int64(999)).Return(nil, store.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "database error (500)",
			userIDParam: "1",
			mockSetup: func(m *store.MockUserStore) {
				m.On("Get", mock.Anything, int64(1)).Return(nil, errors.New("db connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid ID in URL (400)",
			userIDParam:    "abc",
			mockSetup:      func(m *store.MockUserStore) {}, // bd will not be called
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialize the mock user store and set up the expectations
			mockUsers := new(store.MockUserStore)
			tt.mockSetup(mockUsers)

			app := &application{
				store:  store.Storage{Users: mockUsers},
				logger: zap.NewNop().Sugar(),
			}

			// empty next handler to test the middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, err := getTargetUserFromCtx(r)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				w.WriteHeader(http.StatusOK)
			})

			// set up the router with the middleware and the next handler
			r := chi.NewRouter()
			r.With(app.userContextMiddleware).Get("/users/{userID}", nextHandler)

			// set up the request to the router
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userIDParam, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			// check the response code and that the mock was called
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockUsers.AssertExpectations(t)
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
