package ratelimitx

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

//go:embed scripts/sliding_window.lua
var slidingWindowScript string

//go:embed scripts/token_bucket.lua
var tokenBucketScript string

//go:embed scripts/leaky_bucket.lua
var leakyBucketScript string

func (l *Limiter) allowRedis(ctx context.Context, rdb redis.Cmdable, algorithm Algorithm, key string, limit int64, window time.Duration) (bool, int64, error) {
	now := time.Now()
	windowMs := window.Milliseconds()
	if windowMs <= 0 {
		windowMs = 1
	}

	args := []interface{}{now.UnixMilli(), windowMs, limit}
	script := scriptForAlgorithm(algorithm)
	if algorithm == AlgorithmSlidingWindow {
		seq := atomic.AddUint64(&l.seq, 1)
		args = append(args, strconv.FormatInt(now.UnixNano(), 10)+"-"+strconv.FormatUint(seq, 10))
	}

	res, err := rdb.Eval(ctx, script, []string{key}, args...).Result()
	if err != nil {
		return false, 0, err
	}
	return parseRedisResult(res)
}

func scriptForAlgorithm(algorithm Algorithm) string {
	switch algorithm {
	case AlgorithmTokenBucket:
		return tokenBucketScript
	case AlgorithmLeakyBucket:
		return leakyBucketScript
	default:
		return slidingWindowScript
	}
}

func parseRedisResult(res interface{}) (bool, int64, error) {
	values, ok := res.([]interface{})
	if !ok || len(values) < 2 {
		return false, 0, fmt.Errorf("unexpected rate limit redis result: %T", res)
	}
	allowed, err := redisInt64(values[0])
	if err != nil {
		return false, 0, err
	}
	remaining, err := redisInt64(values[1])
	if err != nil {
		return false, 0, err
	}
	return allowed == 1, remaining, nil
}

func redisInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case string:
		return strconv.ParseInt(n, 10, 64)
	case []byte:
		return strconv.ParseInt(string(n), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected rate limit redis value: %T", v)
	}
}
