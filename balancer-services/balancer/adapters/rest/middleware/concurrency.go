package middleware

import (
	"log/slog"
	"net/http"
)

func Concurrency(log *slog.Logger, next http.HandlerFunc, limit int64) http.HandlerFunc {
	limiter := make(chan struct{}, limit)
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case limiter <- struct{}{}:
			next.ServeHTTP(w, r)
			<-limiter
		default:
			log.Warn("concurrency limit exceeded",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"current_concurrency", len(limiter),
				"max_concurrency", limit,
			)
			http.Error(w, "try later", http.StatusServiceUnavailable)
		}
	}
}
