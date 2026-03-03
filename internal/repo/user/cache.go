package user

import (
	"context"
	"fmt"
	"nurture/internal/constant"
	"nurture/internal/global"
	"time"

	"github.com/go-redis/redis/v8"
)

func KeyUserProfile(userID string) string {
	return fmt.Sprintf(constant.USER_PROFILE_KEY, userID)
}

func KeyPartner(userID string) string {
	return fmt.Sprintf(constant.USER_PARTNER_KEY, userID)
}

func KeyFollowing(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, size)
}

func KeyFollowers(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, size)
}

func CacheGet(ctx context.Context, key string) (string, bool, error) {
	if global.RDB == nil {
		return "", false, nil
	}
	v, err := global.RDB.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return v, true, nil
}

func CacheSetEX(ctx context.Context, key string, value string, ttl time.Duration) error {
	if global.RDB == nil {
		return nil
	}
	return global.RDB.SetEX(ctx, key, value, ttl).Err()
}

func CacheDel(ctx context.Context, key string) error {
	if global.RDB == nil {
		return nil
	}
	return global.RDB.Del(ctx, key).Err()
}
