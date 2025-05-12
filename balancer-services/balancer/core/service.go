package core

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
	log      *slog.Logger
}

func NewServerPool(log *slog.Logger) *ServerPool {
	return &ServerPool{backends: make([]*Backend, 0), current: 0, log: log}
}

func (s *ServerPool) AddBackand(address string) {
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	backendURL, _ := url.Parse(address)

	reverseProxy := createProxyWithRetry(backendURL, s.log)
	s.backends = append(s.backends, &Backend{URL: backendURL, ReverseProxy: reverseProxy})
}

func createProxyWithRetry(target *url.URL, log *slog.Logger) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		log.Info("Forwarding request to backend",
			"method", req.Method,
			"path", req.URL.Path,
			"backend", target.Host,
			"client_ip", req.RemoteAddr,
		)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		retries := getRetryCount(r.Context())

		if isConnectionError(err) && retries < 3 {
			log.Error(
				"Retrying request to backend",
				"attempt", retries+1,
				"backend", target.Host,
				"error", err,
			)

			<-time.After(time.Duration(100*(retries+1)) * time.Millisecond)
			ctx := context.WithValue(r.Context(), retryKey{}, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		log.Error(
			"Proxy request failed after retries",
			"retries", retries,
			"error", err,
			"status", "final_error",
		)
		w.WriteHeader(http.StatusBadGateway)
	}

	return proxy
}

func isConnectionError(err error) bool {
	return strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "timeout")
}

type retryKey struct{}

func getRetryCount(ctx context.Context) int {
	if count, ok := ctx.Value(retryKey{}).(int); ok {
		return count
	}
	return 0
}
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)

		if isBackendAlive(s.backends[idx].URL, s.log) {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func isBackendAlive(u *url.URL, log *slog.Logger) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Error("error", "Site unreachable, error: ", err)
		return false
	}
	_ = conn.Close()
	return true
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL, s.log)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		s.log.Info("backend status", "url", b.URL, "status", status)
	}
}
