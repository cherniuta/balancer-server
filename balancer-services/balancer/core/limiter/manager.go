package limiter

import (
	"context"
	"sync"
	"time"
)

type ClientLimiter struct {
	buckets     map[string]*TokenBucketLimiter
	defaultCap  int
	defaultRate time.Duration
	rules       map[string]struct { // Кастомные правила
		limit  int
		period time.Duration
	}
	mu  sync.RWMutex
	ctx context.Context
}

func NewClientLimiter(ctx context.Context, defaultCap int, defaultRate time.Duration) *ClientLimiter {
	return &ClientLimiter{
		buckets:     make(map[string]*TokenBucketLimiter),
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
		rules: make(map[string]struct {
			limit  int
			period time.Duration
		}),
		ctx: ctx,
	}
}

func (cl *ClientLimiter) SetRule(clientID string, limit int, period time.Duration) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.rules[clientID] = struct {
		limit  int
		period time.Duration
	}{limit, period}
}

func (cl *ClientLimiter) Allow(clientID string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	limiter, exists := cl.buckets[clientID]
	if !exists {
		limit, period := cl.defaultCap, cl.defaultRate
		if rule, ok := cl.rules[clientID]; ok {
			limit, period = rule.limit, rule.period
		}
		limiter = NewTokenBucketLimiter(cl.ctx, limit, period)
		cl.buckets[clientID] = limiter
	}

	return limiter.Allow()
}
