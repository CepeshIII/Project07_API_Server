package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/CepeshIII/Project07_API_Server/internal/store"
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
				user := getTargetUserFromCtx(r)
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
