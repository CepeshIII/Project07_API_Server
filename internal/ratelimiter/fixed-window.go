package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]int),
	}
}

// This method tracking request owner ip and allow it if limit isn't reached
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
