package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	userconstant "nurture/internal/user/constant"
	"time"

	"github.com/go-redis/redis/v8"
)

func ProfileKey(userID string) string {
	return fmt.Sprintf(userconstant.ProfileKey, userID)
}

func PartnerKey(userID string) string {
	return fmt.Sprintf(userconstant.PartnerKey, userID)
}

func FollowingKey(userID string, page, size int) string {
	return fmt.Sprintf(userconstant.FollowingKey, userID, page, size)
}

func FollowersKey(userID string, page, size int) string {
	return fmt.Sprintf(userconstant.FollowersKey, userID, page, size)
}

func FollowingPattern(userID string) string {
	return fmt.Sprintf("user:following:%s:*", userID)
}

func FollowersPattern(userID string) string {
	return fmt.Sprintf("user:followers:%s:*", userID)
}

func TagPrefKey(userID string) string {
	return fmt.Sprintf(userconstant.TagPrefKey, userID)
}

func GetString(ctx context.Context, rdb redis.Cmdable, key string) (string, bool, error) {
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
	s, ok, err := GetString(ctx, rdb, key)
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

func ScanDel(ctx context.Context, rdb redis.Cmdable, pattern string, count int64) error {
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
