package httputils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// 1. Перевіряємо Cloudflare (найвищий пріоритет, якщо використовується Cloudflare)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// 2. Перевіряємо X-Forwarded-For (може містити ланцюжок IP: "203.0.113.195, 70.41.3.18, 150.172.238.178")
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Перший IP у списку — це оригінальний клієнт
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	// 3. Перевіряємо X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 4. Якщо заголовків немає, беремо IP з прямого TCP-з'єднання
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
