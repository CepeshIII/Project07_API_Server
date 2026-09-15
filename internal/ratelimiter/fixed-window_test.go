package ratelimiter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedWindowRateLimiterAllow(t *testing.T) {
	t.Run("check if limiter does not allow request after crossing the limit", func(t *testing.T) {
		limit := 100
		requestsCount := 150

		window := 1 * time.Second
		fwRateLimiter := NewFixedWindowRateLimiter(limit, window)

		ip := "192.168.1.1"

		countOfAllow := 0
		countOfNotAllow := 0

		for range requestsCount {
			ok, _ := fwRateLimiter.Allow(ip)
			if ok {
				countOfAllow++
			} else {
				countOfNotAllow++
			}

		}

		assert.Equal(t, limit, countOfAllow)
		assert.Equal(t, requestsCount-limit, countOfNotAllow)
	})

	t.Run("check if limiter allows request after resetting", func(t *testing.T) {
		limit := 100
		requestsCount := 180

		window := 1 * time.Second
		fwRateLimiter := NewFixedWindowRateLimiter(limit, window)

		ip := "192.168.1.1"

		countOfAllow := 0
		countOfNotAllow := 0

		halfRequests := requestsCount / 2

		for range halfRequests {
			ok, _ := fwRateLimiter.Allow(ip)
			if ok {
				countOfAllow++
			} else {
				countOfNotAllow++
			}
		}

		time.Sleep(window * 2)

		for range halfRequests {
			ok, _ := fwRateLimiter.Allow(ip)
			if ok {
				countOfAllow++
			} else {
				countOfNotAllow++
			}
		}

		assert.Equal(t, requestsCount, countOfAllow)
		assert.Equal(t, 0, countOfNotAllow)
	})

	t.Run("check if limiter isolates different IPs", func(t *testing.T) {
		limit := 2
		window := 1 * time.Second
		fwRateLimiter := NewFixedWindowRateLimiter(limit, window)

		ip1 := "192.168.1.1"
		ip2 := "192.168.1.2"

		// Exhaust limit for IP1
		ok1, _ := fwRateLimiter.Allow(ip1)
		ok2, _ := fwRateLimiter.Allow(ip1)
		ok3, _ := fwRateLimiter.Allow(ip1) // Should fail

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.False(t, ok3)

		// IP2 should still be allowed
		ip2Ok1, _ := fwRateLimiter.Allow(ip2)
		assert.True(t, ip2Ok1)
	})

	t.Run("check thread safety under concurrent load", func(t *testing.T) {
		limit := 50
		window := 2 * time.Second
		fwRateLimiter := NewFixedWindowRateLimiter(limit, window)

		ip := "192.168.1.1"
		concurrency := 100

		var wg sync.WaitGroup
		wg.Add(concurrency)

		var allowedCount int
		var mu sync.Mutex

		for range concurrency {
			go func() {
				defer wg.Done()
				ok, _ := fwRateLimiter.Allow(ip)
				if ok {
					mu.Lock()
					allowedCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, limit, allowedCount)
	})
}
