package jwtx

import (
	"context"
	"sync"
	"time"
)

type memoryTokenItem struct {
	session  RTokenSession
	expireAt time.Time
}

type memoryMarker struct {
	expireAt time.Time
}

type MemoryTokenStore struct {
	mu              sync.Mutex
	activeRTokens   map[string]memoryTokenItem
	usedRTokens     map[string]memoryMarker
	atokenBlacklist map[string]memoryMarker
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		activeRTokens:   make(map[string]memoryTokenItem),
		usedRTokens:     make(map[string]memoryMarker),
		atokenBlacklist: make(map[string]memoryMarker),
	}
}

func (s *MemoryTokenStore) SetActiveRToken(ctx context.Context, hash string, session RTokenSession, ttl time.Duration) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl <= 0 {
		return ErrTokenExpired
	}
	s.activeRTokens[hash] = memoryTokenItem{session: session, expireAt: time.Now().Add(ttl)}
	return nil
}

func (s *MemoryTokenStore) RotateRToken(ctx context.Context, oldHash, newHash string, session RTokenSession, activeTTL, usedTTL time.Duration) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.cleanupLocked(now)
	if _, ok := s.activeRTokens[oldHash]; !ok {
		if _, used := s.usedRTokens[oldHash]; used {
			return ErrRTokenReplay
		}
		return ErrTokenInvalid
	}
	delete(s.activeRTokens, oldHash)
	s.activeRTokens[newHash] = memoryTokenItem{session: session, expireAt: now.Add(activeTTL)}
	if usedTTL > 0 {
		s.usedRTokens[oldHash] = memoryMarker{expireAt: now.Add(usedTTL)}
	}
	return nil
}

func (s *MemoryTokenStore) DeleteActiveRToken(ctx context.Context, hash string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.activeRTokens, hash)
	return nil
}

func (s *MemoryTokenStore) MarkRTokenUsed(ctx context.Context, hash string, ttl time.Duration) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl > 0 {
		s.usedRTokens[hash] = memoryMarker{expireAt: time.Now().Add(ttl)}
	}
	return nil
}

func (s *MemoryTokenStore) IsRTokenUsed(ctx context.Context, hash string) (bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now())
	_, ok := s.usedRTokens[hash]
	return ok, nil
}

func (s *MemoryTokenStore) BlacklistAToken(ctx context.Context, hash string, ttl time.Duration) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl > 0 {
		s.atokenBlacklist[hash] = memoryMarker{expireAt: time.Now().Add(ttl)}
	}
	return nil
}

func (s *MemoryTokenStore) IsATokenBlacklisted(ctx context.Context, hash string) (bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now())
	_, ok := s.atokenBlacklist[hash]
	return ok, nil
}

func (s *MemoryTokenStore) cleanupLocked(now time.Time) {
	for hash, item := range s.activeRTokens {
		if !item.expireAt.IsZero() && !item.expireAt.After(now) {
			delete(s.activeRTokens, hash)
		}
	}
	for hash, item := range s.usedRTokens {
		if !item.expireAt.IsZero() && !item.expireAt.After(now) {
			delete(s.usedRTokens, hash)
		}
	}
	for hash, item := range s.atokenBlacklist {
		if !item.expireAt.IsZero() && !item.expireAt.After(now) {
			delete(s.atokenBlacklist, hash)
		}
	}
}

var _ TokenStore = (*MemoryTokenStore)(nil)
