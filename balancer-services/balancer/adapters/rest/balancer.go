package rest

import (
	"log/slog"
	"net/http"
	"test-assignment/balancer/core"
)

func HandleRequest(log *slog.Logger, serverPool core.Balancer) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		next := serverPool.GetNextPeer()

		if next != nil {
			next.ReverseProxy.ServeHTTP(writer, request)
			return
		}
		http.Error(writer, "Service not available", http.StatusServiceUnavailable)
		log.Error(
			"Backend failure",
			slog.Group("context",
				"error", "no available servers",
				"action", "retrying in 5s",
				"backends_count", 0,
			),
		)
	}
}
