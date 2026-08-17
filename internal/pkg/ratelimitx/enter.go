package ratelimitx

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Algorithm string

const (
	AlgorithmSlidingWindow Algorithm = "sliding_window"
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmLeakyBucket   Algorithm = "leaky_bucket"
)

type Limiter struct {
	mu        sync.RWMutex
	rdb       redis.Cmdable
	algorithm Algorithm
	mem       *memLimiter
	seq       uint64
}

func NewLimiter(rdb redis.Cmdable) *Limiter {
	return NewLimiterWithAlgorithm(rdb, AlgorithmSlidingWindow)
}

func NewLimiterWithAlgorithm(rdb redis.Cmdable, algorithm Algorithm) *Limiter {
	return &Limiter{
		rdb:       rdb,
		algorithm: normalizeAlgorithm(algorithm),
		mem:       newMemLimiter(),
	}
}

func (l *Limiter) SetRedis(rdb redis.Cmdable) {
	l.mu.Lock()
	l.rdb = rdb
	l.mu.Unlock()
}

func (l *Limiter) SetAlgorithm(algorithm Algorithm) {
	l.mu.Lock()
	l.algorithm = normalizeAlgorithm(algorithm)
	l.mu.Unlock()
}

func (l *Limiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	l.mu.RLock()
	algorithm := l.algorithm
	l.mu.RUnlock()
	return l.AllowWithAlgorithm(ctx, algorithm, key, limit, window)
}

func (l *Limiter) AllowWithAlgorithm(ctx context.Context, algorithm Algorithm, key string, limit int64, window time.Duration) (bool, int64, error) {
	if limit <= 0 {
		return true, 0, nil
	}
	window = normalizeWindow(window)
	algorithm = normalizeAlgorithm(algorithm)

	l.mu.RLock()
	rdb := l.rdb
	l.mu.RUnlock()
	if rdb == nil {
		allowed, remaining := l.mem.allow(algorithm, key, limit, window)
		return allowed, remaining, nil
	}

	allowed, remaining, err := l.allowRedis(ctx, rdb, algorithm, key, limit, window)
	if err != nil {
		allowed, remaining = l.mem.allow(algorithm, key, limit, window)
		return allowed, remaining, nil
	}
	return allowed, remaining, nil
}

func normalizeAlgorithm(algorithm Algorithm) Algorithm {
	switch algorithm {
	case AlgorithmTokenBucket, AlgorithmLeakyBucket, AlgorithmSlidingWindow:
		return algorithm
	default:
		return AlgorithmSlidingWindow
	}
}

func normalizeWindow(window time.Duration) time.Duration {
	if window <= 0 {
		return time.Second
	}
	return window
}
