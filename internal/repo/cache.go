package repo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

func getCacheString(ctx context.Context, rdb redis.Cmdable, key string) (string, bool, error) {
	if rdb == nil {
		return "", false, nil
	}
	v, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", false, nil
		}
		return "", false, err
	}
	return v, true, nil
}

func setCacheEX(ctx context.Context, rdb redis.Cmdable, key string, value string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	return rdb.SetEX(ctx, key, value, ttl).Err()
}

func delCache(ctx context.Context, rdb redis.Cmdable, keys ...string) error {
	if rdb == nil {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}

func getCacheJSON[T any](ctx context.Context, rdb redis.Cmdable, key string, dst *T) (bool, error) {
	s, ok, err := getCacheString(ctx, rdb, key)
	if err != nil || !ok {
		return ok, err
	}
	if s == "" {
		return false, nil
	}
	if err := json.Unmarshal([]byte(s), dst); err != nil {
		return false, nil
	}
	return true, nil
}

func setCacheJSON(ctx context.Context, rdb redis.Cmdable, key string, value any, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return setCacheEX(ctx, rdb, key, string(b), ttl)
}

func scanDelCache(ctx context.Context, rdb redis.Cmdable, pattern string, count int64) error {
	if rdb == nil {
		return nil
	}
	if count <= 0 {
		count = 100
	}
	iter := rdb.Scan(ctx, 0, pattern, count).Iterator()
	for iter.Next(ctx) {
		_ = rdb.Del(ctx, iter.Val()).Err()
	}
	return iter.Err()
}
