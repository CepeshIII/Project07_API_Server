package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CepeshIII/Project07_API_Server/internal/ratelimiter"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRateLimiterIntegration(t *testing.T) {
	ip := "192.168.1.1"
	limit := 1
	window := 2 * time.Second

	// Use the real limiter instead of the mock
	limiter := ratelimiter.NewFixedWindowRateLimiter(limit, window)

	app := &application{
		rateLimiter: limiter, // Ensure your app struct uses an interface type: RateLimiter
		logger:      zap.NewNop().Sugar(),
	}

	r := chi.NewRouter()
	r.With(app.rateLimiterMiddleWare).Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1. First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = ip
	rr1 := httptest.NewRecorder()
	r.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// 2. Second request should be rate-limited
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = ip
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
	assert.NotEmpty(t, rr2.Header().Get("Retry-After"))
}
