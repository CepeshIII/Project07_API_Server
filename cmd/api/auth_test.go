package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CepeshIII/Project07_API_Server/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAccessTokenMiddleware(t *testing.T) {
	type test struct {
		name           string
		mockSetup      func(m *auth.MockAuthenticator)
		cookieSetup    func(r *http.Request)
		expectedStatus int
	}

	tests := []test{
		{
			name: "Valid access token",
			mockSetup: func(m *auth.MockAuthenticator) {
				m.Mock.On("ValidateToken", "valid-token").Return(&auth.CustomClaims{UserID: 1}, nil)
			},
			cookieSetup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: accessCookieName, Value: "valid-token"})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid access token",
			mockSetup: func(m *auth.MockAuthenticator) {
				m.Mock.On("ValidateToken", "invalid-token").Return((*auth.CustomClaims)(nil), auth.ErrInvalidToken)
			},
			cookieSetup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: accessCookieName, Value: "invalid-token"})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Cookie not found",
			mockSetup: func(m *auth.MockAuthenticator) {
				// No need to set up the mock for this case
			},
			cookieSetup: func(r *http.Request) {
				// Do not add the cookie
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAuth := new(auth.MockAuthenticator)
			tt.mockSetup(mockAuth)

			app := &application{
				auth:   mockAuth,
				logger: zap.NewNop().Sugar(),
			}

			// empty next handler to test the middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, err := getAuthUserIDFromCtx(r)
				assert.NotNil(t, user)
				assert.NoError(t, err)
				w.WriteHeader(http.StatusOK)
			})

			// set up the router with the middleware and the next handler
			r := chi.NewRouter()
			r.With(app.accessTokenMiddleware).Get("/", nextHandler)

			// set up the request and recorder
			req := httptest.NewRequest("GET", "/", nil)
			tt.cookieSetup(req)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req) // call the router with the request

			// check the response code and that the mock was called
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockAuth.AssertExpectations(t)
		})
	}

}
