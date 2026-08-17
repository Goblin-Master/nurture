package ratelimitx

import "time"

func (m *memLimiter) allowSlidingWindow(key string, limit int64, window time.Duration, now time.Time) (bool, int64) {
	cutoff := now.Add(-window).UnixNano()
	items := m.windows[key]
	start := 0
	for start < len(items) && items[start] <= cutoff {
		start++
	}
	if start > 0 {
		copy(items, items[start:])
		items = items[:len(items)-start]
	}
	if int64(len(items)) >= limit {
		m.windows[key] = items
		return false, 0
	}
	items = append(items, now.UnixNano())
	m.windows[key] = items
	return true, limit - int64(len(items))
}
