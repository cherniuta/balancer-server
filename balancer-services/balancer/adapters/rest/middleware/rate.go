package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"test-assignment/balancer/core/limiter"
)

func Rate(log *slog.Logger, limiter *limiter.ClientLimiter) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := getClientID(r)

			if !limiter.Allow(clientID) {
				log.Warn("rate limit exceeded",
					"client_id", clientID,
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientID(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return "api_key_" + key
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return "ip_" + ip
}
