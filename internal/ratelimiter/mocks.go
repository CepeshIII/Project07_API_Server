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
	args := rl.Mock.Called(ip)
	return args.Bool(0), args.Get(1).(time.Duration)
}
