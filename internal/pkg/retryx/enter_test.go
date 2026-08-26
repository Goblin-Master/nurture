package retryx

import (
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name     string
		base     time.Duration
		attempt  int64
		maxDelay time.Duration
		want     time.Duration
	}{
		{name: "first attempt uses base", base: time.Second, attempt: 1, maxDelay: 32 * time.Second, want: time.Second},
		{name: "second attempt doubles base", base: time.Second, attempt: 2, maxDelay: 32 * time.Second, want: 2 * time.Second},
		{name: "later attempt grows exponentially", base: time.Second, attempt: 6, maxDelay: 32 * time.Second, want: 32 * time.Second},
		{name: "attempt below one is normalized", base: time.Second, attempt: 0, maxDelay: 32 * time.Second, want: time.Second},
		{name: "caps by max delay", base: time.Second, attempt: 7, maxDelay: 32 * time.Second, want: 32 * time.Second},
		{name: "zero max delay disables cap", base: time.Second, attempt: 7, want: 64 * time.Second},
		{name: "zero base returns zero", attempt: 2, maxDelay: 32 * time.Second, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExponentialBackoff(tt.base, tt.attempt, tt.maxDelay)
			if got != tt.want {
				t.Fatalf("ExponentialBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}
