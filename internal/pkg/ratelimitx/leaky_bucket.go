package ratelimitx

import (
	"math"
	"time"
)

type memLeakyBucket struct {
	level float64
	last  time.Time
}

func (m *memLimiter) allowLeakyBucket(key string, limit int64, window time.Duration, now time.Time) (bool, int64) {
	capacity := float64(limit)
	item, ok := m.leaks[key]
	if !ok {
		item = memLeakyBucket{last: now}
	} else {
		elapsed := now.Sub(item.last)
		if elapsed < 0 {
			elapsed = 0
		}
		item.level = math.Max(0, item.level-elapsed.Seconds()*capacity/window.Seconds())
		item.last = now
	}

	if item.level+1 > capacity {
		m.leaks[key] = item
		return false, floorRemaining(capacity - item.level)
	}
	item.level++
	m.leaks[key] = item
	return true, floorRemaining(capacity - item.level)
}
