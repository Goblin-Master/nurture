package jwtx

import (
	"context"
	"time"
)

type RTokenSession struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
}

type TokenStore interface {
	SetActiveRToken(ctx context.Context, hash string, session RTokenSession, ttl time.Duration) error
	RotateRToken(ctx context.Context, oldHash, newHash string, session RTokenSession, activeTTL, usedTTL time.Duration) error
	DeleteActiveRToken(ctx context.Context, hash string) error
	MarkRTokenUsed(ctx context.Context, hash string, ttl time.Duration) error
	IsRTokenUsed(ctx context.Context, hash string) (bool, error)
	BlacklistAToken(ctx context.Context, hash string, ttl time.Duration) error
	IsATokenBlacklisted(ctx context.Context, hash string) (bool, error)
}
