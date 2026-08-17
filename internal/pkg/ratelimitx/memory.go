package ratelimitx

import (
	"math"
	"sync"
	"time"
)

type memLimiter struct {
	mu      sync.Mutex
	windows map[string][]int64
	tokens  map[string]memTokenBucket
	leaks   map[string]memLeakyBucket
}

func newMemLimiter() *memLimiter {
	return &memLimiter{
		windows: make(map[string][]int64),
		tokens:  make(map[string]memTokenBucket),
		leaks:   make(map[string]memLeakyBucket),
	}
}

func (m *memLimiter) allow(algorithm Algorithm, key string, limit int64, window time.Duration) (bool, int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	switch algorithm {
	case AlgorithmTokenBucket:
		return m.allowTokenBucket(key, limit, window, now)
	case AlgorithmLeakyBucket:
		return m.allowLeakyBucket(key, limit, window, now)
	default:
		return m.allowSlidingWindow(key, limit, window, now)
	}
}

func floorRemaining(v float64) int64 {
	if v <= 0 {
		return 0
	}
	return int64(math.Floor(v))
}
