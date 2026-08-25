package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	babyconstant "nurture/internal/baby/constant"
	"time"

	"github.com/go-redis/redis/v8"
)

func InfoKey(babyID, userID string) string {
	return fmt.Sprintf(babyconstant.InfoKey, babyID, userID)
}

func VaccineListKey(babyID string) string {
	return fmt.Sprintf(babyconstant.VaccineListKey, babyID)
}

func LatestGrowthKey(babyID string) string {
	return fmt.Sprintf(babyconstant.LatestGrowthKey, babyID)
}

func InfoPattern(babyID string) string {
	return fmt.Sprintf("baby:info:%s:*", babyID)
}

func getString(ctx context.Context, rdb redis.Cmdable, key string) (string, bool, error) {
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

func SetEX(ctx context.Context, rdb redis.Cmdable, key string, value string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	return rdb.SetEX(ctx, key, value, ttl).Err()
}

func Del(ctx context.Context, rdb redis.Cmdable, keys ...string) error {
	if rdb == nil || len(keys) == 0 {
		return nil
	}
	return rdb.Del(ctx, keys...).Err()
}

func GetJSON[T any](ctx context.Context, rdb redis.Cmdable, key string, dst *T) (bool, error) {
	s, ok, err := getString(ctx, rdb, key)
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

func SetJSON(ctx context.Context, rdb redis.Cmdable, key string, value any, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return SetEX(ctx, rdb, key, string(b), ttl)
}
