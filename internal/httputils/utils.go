// Package httputils provides HTTP helpers and utility functions for extracting headers,
// client information, and request processing.
package httputils

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts the real client IP address from request headers (Cloudflare, X-Forwarded-For, X-Real-IP)
// or falls back to the remote TCP connection address.
func GetClientIP(r *http.Request) string {
	// 1. Check Cloudflare (highest priority if Cloudflare proxy is used)
	if cfIP := r.Header.Get("Cf-Connecting-Ip"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// 2. Check X-Forwarded-For (may contain an IP chain: "203.0.113.195, 70.41.3.18, 150.172.238.178")
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// The first IP in the list is the original client
		ips := strings.Split(xff, ",")

		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	// 3. Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 4. Fallback to direct TCP connection address if no proxy headers are present
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
