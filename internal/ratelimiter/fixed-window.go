package ratelimiter

import (
	"sync"
	"time"
)

// FixedWindowRateLimiter implements a fixed window rate limiter using an in-memory map.
type FixedWindowRateLimiter struct {
	sync.RWMutex

	clients map[string]int
	limit   int
	window  time.Duration
}

// NewFixedWindowRateLimiter creates and returns a new FixedWindowRateLimiter instance.
func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]int),
	}
}

// Allow tracks client requests by IP and permits them if the limit has not been reached.
func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.Lock()
	defer rl.Unlock()

	count, exist := rl.clients[ip]

	if !exist {
		// First request for this IP in the current window
		rl.clients[ip] = 1
		go rl.resetCount(ip)

		return true, 0
	}

	if count < rl.limit {
		rl.clients[ip]++

		return true, 0
	}

	return false, rl.window
}

func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)

	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}
