package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	postconstant "nurture/internal/post/constant"
	"time"

	"github.com/go-redis/redis/v8"
)

func HotDetailKey(postID, userID string) string {
	return fmt.Sprintf(postconstant.HotDetailKey, postID) + ":" + userID
}

func HotListKey(userID string, page, pageSize int) string {
	return fmt.Sprintf(postconstant.HotListKey, page, pageSize) + ":" + userID
}

func HotListByTagKey(userID, tagID string, page, pageSize int) string {
	return fmt.Sprintf(postconstant.HotListByTagKey, tagID, page, pageSize) + ":" + userID
}

func HotCommentsKey(postID, userID string, page, pageSize int) string {
	return fmt.Sprintf(postconstant.HotCommentsKey, postID, userID, page, pageSize)
}

func CommentHotRepliesKey(commentID, userID string, page, pageSize int) string {
	return fmt.Sprintf(postconstant.CommentHotRepliesKey, commentID, userID, page, pageSize)
}

func UserTagPrefKey(userID string) string {
	return fmt.Sprintf(postconstant.UserTagPrefKey, userID)
}

func HotListPattern(userID string) string {
	return fmt.Sprintf("post:list:hot:*:*:%s", userID)
}

func HotCommentsPattern(postID, userID string) string {
	return fmt.Sprintf("post:comments:hot:%s:%s:*:*", postID, userID)
}

func CommentHotRepliesPattern(commentID, userID string) string {
	return fmt.Sprintf("comment:replies:hot:%s:%s:*:*", commentID, userID)
}

func HotDetailPattern(postID string) string {
	return fmt.Sprintf("post:detail:hot:%s:*", postID)
}

func HotCommentsAllUsersPattern(postID string) string {
	return fmt.Sprintf("post:comments:hot:%s:*", postID)
}

func HotListAllPattern() string {
	return "post:list:hot:*:*:*"
}

func HotListByTagAllPattern() string {
	return "post:list:hot:tag:*:*:*:*"
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
