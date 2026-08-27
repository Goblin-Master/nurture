package jwtx

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	atokenBlacklistKey = "auth:atoken:blacklist:%s"
	rtokenActiveKey    = "auth:rtoken:active:%s"
	rtokenUsedKey      = "auth:rtoken:used:%s"
)

//go:embed scripts/rotate_rtoken.lua
var rotateRTokenScript string

type RedisTokenStore struct {
	rdb redis.Cmdable
}

func NewRedisTokenStore(rdb redis.Cmdable) *RedisTokenStore {
	return &RedisTokenStore{rdb: rdb}
}

func (s *RedisTokenStore) SetActiveRToken(ctx context.Context, hash string, session RTokenSession, ttl time.Duration) error {
	if s.rdb == nil {
		return ErrTokenStore
	}
	if ttl <= 0 {
		return ErrTokenExpired
	}
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, fmt.Sprintf(rtokenActiveKey, hash), data, ttl).Err()
}

func (s *RedisTokenStore) RotateRToken(ctx context.Context, oldHash, newHash string, session RTokenSession, activeTTL, usedTTL time.Duration) error {
	if s.rdb == nil {
		return ErrTokenStore
	}
	if activeTTL <= 0 {
		return ErrTokenExpired
	}
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	res, err := s.rdb.Eval(
		ctx,
		rotateRTokenScript,
		[]string{
			fmt.Sprintf(rtokenActiveKey, oldHash),
			fmt.Sprintf(rtokenActiveKey, newHash),
			fmt.Sprintf(rtokenUsedKey, oldHash),
		},
		data,
		millis(activeTTL),
		millis(usedTTL),
	).Int()
	if err != nil {
		return err
	}
	switch res {
	case 1:
		return nil
	case -1:
		return ErrRTokenReplay
	default:
		return ErrTokenInvalid
	}
}

func (s *RedisTokenStore) DeleteActiveRToken(ctx context.Context, hash string) error {
	if s.rdb == nil {
		return ErrTokenStore
	}
	return s.rdb.Del(ctx, fmt.Sprintf(rtokenActiveKey, hash)).Err()
}

func (s *RedisTokenStore) MarkRTokenUsed(ctx context.Context, hash string, ttl time.Duration) error {
	if s.rdb == nil {
		return ErrTokenStore
	}
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, fmt.Sprintf(rtokenUsedKey, hash), "1", ttl).Err()
}

func (s *RedisTokenStore) IsRTokenUsed(ctx context.Context, hash string) (bool, error) {
	if s.rdb == nil {
		return false, ErrTokenStore
	}
	count, err := s.rdb.Exists(ctx, fmt.Sprintf(rtokenUsedKey, hash)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (s *RedisTokenStore) BlacklistAToken(ctx context.Context, hash string, ttl time.Duration) error {
	if s.rdb == nil {
		return ErrTokenStore
	}
	if ttl <= 0 {
		return nil
	}
	return s.rdb.Set(ctx, fmt.Sprintf(atokenBlacklistKey, hash), "1", ttl).Err()
}

func (s *RedisTokenStore) IsATokenBlacklisted(ctx context.Context, hash string) (bool, error) {
	if s.rdb == nil {
		return false, ErrTokenStore
	}
	count, err := s.rdb.Exists(ctx, fmt.Sprintf(atokenBlacklistKey, hash)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func millis(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return int64(d / time.Millisecond)
}

var _ TokenStore = (*RedisTokenStore)(nil)
