// Package ratelimiter provides rate limiting interfaces and configuration structures.
package ratelimiter

import "time"

// Limiter defines the interface for checking whether a given IP address is allowed to make a request.
type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

// Config holds the configuration parameters for rate limiting.
type Config struct {
	RequestsPerTimeFrame int
	TimeFrame            time.Duration
	Enabled              bool
}
