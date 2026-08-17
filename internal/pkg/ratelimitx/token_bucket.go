package ratelimitx

import (
	"math"
	"time"
)

type memTokenBucket struct {
	tokens float64
	last   time.Time
}

func (m *memLimiter) allowTokenBucket(key string, limit int64, window time.Duration, now time.Time) (bool, int64) {
	capacity := float64(limit)
	item, ok := m.tokens[key]
	if !ok {
		item = memTokenBucket{
			tokens: capacity,
			last:   now,
		}
	} else {
		elapsed := now.Sub(item.last)
		if elapsed < 0 {
			elapsed = 0
		}
		item.tokens = math.Min(capacity, item.tokens+elapsed.Seconds()*capacity/window.Seconds())
		item.last = now
	}

	if item.tokens < 1 {
		m.tokens[key] = item
		return false, floorRemaining(item.tokens)
	}
	item.tokens--
	m.tokens[key] = item
	return true, floorRemaining(item.tokens)
}
