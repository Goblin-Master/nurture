package retryx

import "time"

const maxDuration time.Duration = 1<<63 - 1

// ExponentialBackoff returns base * 2^(attempt-1), capped by maxDelay when maxDelay is positive.
func ExponentialBackoff(base time.Duration, attempt int64, maxDelay time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	if attempt <= 1 {
		return capDelay(base, maxDelay)
	}
	delay := base
	for i := int64(1); i < attempt; i++ {
		if maxDelay > 0 && delay >= maxDelay {
			return maxDelay
		}
		if delay > maxDuration/2 {
			delay = maxDuration
		} else {
			delay *= 2
		}
	}
	return capDelay(delay, maxDelay)
}

func capDelay(delay time.Duration, maxDelay time.Duration) time.Duration {
	if maxDelay > 0 && delay > maxDelay {
		return maxDelay
	}
	return delay
}
