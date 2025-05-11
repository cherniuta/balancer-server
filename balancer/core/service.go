package core

import (
	"log/slog"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
	log      *slog.Logger
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)

		if s.backends[idx].IsAlive() {
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
		s.log.Info("%s [%s]\n", b.URL, status)
	}
}
