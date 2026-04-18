package ratelimitx

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type WindowLimiter struct {
	mu  sync.RWMutex
	rdb redis.Cmdable
	mem *memWindowLimiter
}

func NewWindowLimiter(rdb redis.Cmdable) *WindowLimiter {
	return &WindowLimiter{
		rdb: rdb,
		mem: newMemWindowLimiter(),
	}
}

func (l *WindowLimiter) SetRedis(rdb redis.Cmdable) {
	l.mu.Lock()
	l.rdb = rdb
	l.mu.Unlock()
}

func (l *WindowLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	if limit <= 0 {
		return true, 0, nil
	}
	if window <= 0 {
		window = time.Second
	}
	l.mu.RLock()
	rdb := l.rdb
	l.mu.RUnlock()
	if rdb == nil {
		return l.mem.allow(key, limit, window), 0, nil
	}
	ttl := window
	n, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return l.mem.allow(key, limit, window), 0, nil
	}
	if n == 1 {
		_ = rdb.Expire(ctx, key, ttl).Err()
	}
	if n > limit {
		return false, 0, nil
	}
	return true, limit - n, nil
}

type memWindowLimiter struct {
	mu   sync.Mutex
	data map[string]memWindowItem
}

type memWindowItem struct {
	slot  int64
	count int64
}

func newMemWindowLimiter() *memWindowLimiter {
	return &memWindowLimiter{
		data: make(map[string]memWindowItem),
	}
}

func (m *memWindowLimiter) allow(key string, limit int64, window time.Duration) bool {
	now := time.Now().UnixNano()
	slot := now / window.Nanoseconds()
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.data[key]
	if !ok || item.slot != slot {
		m.data[key] = memWindowItem{slot: slot, count: 1}
		return true
	}
	if item.count >= limit {
		return false
	}
	item.count++
	m.data[key] = item
	return true
}
