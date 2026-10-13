package ratelimiter

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRateLimiter struct {
	mock.Mock
}

func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{}
}

func (rl *MockRateLimiter) Allow(ip string) (bool, time.Duration) {
	args := rl.Called(ip)

	duration, ok := args.Get(1).(time.Duration)
	if !ok {
		return args.Bool(0), 0
	}

	return args.Bool(0), duration
}
