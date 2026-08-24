package ratelimitx

import (
	"context"
	"fmt"
	"nurture/internal/config"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRateLimitDefaultIsSlidingWindow(t *testing.T) {
	limiter := NewLimiter(nil)
	key := testRateLimitKey("default")

	mustAllowDefault(t, limiter, key, 2, 50*time.Millisecond)
	mustAllowDefault(t, limiter, key, 2, 50*time.Millisecond)
	mustDenyDefault(t, limiter, key, 2, 50*time.Millisecond)

	mustAllowDefaultEventually(t, limiter, key, 2, 50*time.Millisecond, time.Second)
}

func TestRateLimitSlidingWindowMemory(t *testing.T) {
	limiter := NewLimiterWithAlgorithm(nil, AlgorithmSlidingWindow)
	key := testRateLimitKey("sliding")

	mustAllow(t, limiter, AlgorithmSlidingWindow, key, 2, 50*time.Millisecond)
	mustAllow(t, limiter, AlgorithmSlidingWindow, key, 2, 50*time.Millisecond)
	mustDeny(t, limiter, AlgorithmSlidingWindow, key, 2, 50*time.Millisecond)

	mustAllowEventually(t, limiter, AlgorithmSlidingWindow, key, 2, 50*time.Millisecond, time.Second)
}

func TestRateLimitTokenBucketMemory(t *testing.T) {
	limiter := NewLimiterWithAlgorithm(nil, AlgorithmTokenBucket)
	key := testRateLimitKey("token")

	mustAllow(t, limiter, AlgorithmTokenBucket, key, 2, 100*time.Millisecond)
	mustAllow(t, limiter, AlgorithmTokenBucket, key, 2, 100*time.Millisecond)
	mustDeny(t, limiter, AlgorithmTokenBucket, key, 2, 100*time.Millisecond)

	mustAllowEventually(t, limiter, AlgorithmTokenBucket, key, 2, 100*time.Millisecond, time.Second)
}

func TestRateLimitLeakyBucketMemory(t *testing.T) {
	limiter := NewLimiterWithAlgorithm(nil, AlgorithmLeakyBucket)
	key := testRateLimitKey("leaky")

	mustAllow(t, limiter, AlgorithmLeakyBucket, key, 2, 100*time.Millisecond)
	mustAllow(t, limiter, AlgorithmLeakyBucket, key, 2, 100*time.Millisecond)
	mustDeny(t, limiter, AlgorithmLeakyBucket, key, 2, 100*time.Millisecond)

	mustAllowEventually(t, limiter, AlgorithmLeakyBucket, key, 2, 100*time.Millisecond, time.Second)
}

func TestRateLimitRedisAlgorithms(t *testing.T) {
	config.LoadConfig()
	if !config.Conf.Redis.Enable {
		t.Skip("skip redis integration test: redis.enable=false")
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.DSN(),
		Username: config.Conf.Redis.UserName,
		Password: config.Conf.Redis.Password,
		DB:       config.Conf.Redis.DB,
	})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}

	cases := []struct {
		name      string
		algorithm Algorithm
		window    time.Duration
	}{
		{
			name:      "sliding_window",
			algorithm: AlgorithmSlidingWindow,
			window:    50 * time.Millisecond,
		},
		{
			name:      "token_bucket",
			algorithm: AlgorithmTokenBucket,
			window:    300 * time.Millisecond,
		},
		{
			name:      "leaky_bucket",
			algorithm: AlgorithmLeakyBucket,
			window:    300 * time.Millisecond,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			limiter := NewLimiterWithAlgorithm(rdb, c.algorithm)
			key := testRateLimitKey(c.name)
			defer rdb.Del(ctx, key)

			mustAllow(t, limiter, c.algorithm, key, 2, c.window)
			mustAllow(t, limiter, c.algorithm, key, 2, c.window)
			mustDeny(t, limiter, c.algorithm, key, 2, c.window)

			mustAllowEventually(t, limiter, c.algorithm, key, 2, c.window, time.Second)
		})
	}
}

func testRateLimitKey(name string) string {
	return fmt.Sprintf("test:rl:%s:%d", name, time.Now().UnixNano())
}

func mustAllow(t *testing.T, limiter *Limiter, algorithm Algorithm, key string, limit int64, window time.Duration) {
	t.Helper()
	ok, remaining, err := limiter.AllowWithAlgorithm(context.Background(), algorithm, key, limit, window)
	if err != nil {
		t.Fatalf("allow returned error: %v", err)
	}
	if !ok {
		t.Fatalf("allow returned false, remaining=%d", remaining)
	}
}

func mustAllowDefault(t *testing.T, limiter *Limiter, key string, limit int64, window time.Duration) {
	t.Helper()
	ok, remaining, err := limiter.Allow(context.Background(), key, limit, window)
	if err != nil {
		t.Fatalf("allow returned error: %v", err)
	}
	if !ok {
		t.Fatalf("allow returned false, remaining=%d", remaining)
	}
}

func mustAllowDefaultEventually(t *testing.T, limiter *Limiter, key string, limit int64, window time.Duration, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		ok, remaining, err := limiter.Allow(context.Background(), key, limit, window)
		if err != nil {
			t.Fatalf("allow returned error: %v", err)
		}
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("allow did not return true before timeout, remaining=%d", remaining)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func mustDenyDefault(t *testing.T, limiter *Limiter, key string, limit int64, window time.Duration) {
	t.Helper()
	ok, remaining, err := limiter.Allow(context.Background(), key, limit, window)
	if err != nil {
		t.Fatalf("allow returned error: %v", err)
	}
	if ok {
		t.Fatalf("allow returned true, remaining=%d", remaining)
	}
}

func mustDeny(t *testing.T, limiter *Limiter, algorithm Algorithm, key string, limit int64, window time.Duration) {
	t.Helper()
	ok, remaining, err := limiter.AllowWithAlgorithm(context.Background(), algorithm, key, limit, window)
	if err != nil {
		t.Fatalf("allow returned error: %v", err)
	}
	if ok {
		t.Fatalf("allow returned true, remaining=%d", remaining)
	}
}

func mustAllowEventually(t *testing.T, limiter *Limiter, algorithm Algorithm, key string, limit int64, window time.Duration, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		ok, remaining, err := limiter.AllowWithAlgorithm(context.Background(), algorithm, key, limit, window)
		if err != nil {
			t.Fatalf("allow returned error: %v", err)
		}
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("allow did not return true before timeout, remaining=%d", remaining)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
