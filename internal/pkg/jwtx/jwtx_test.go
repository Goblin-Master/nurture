package jwtx

import (
	"context"
	"errors"
	"nurture/internal/config"
	"testing"
)

func TestTokenPairParsesOnlyExpectedTypes(t *testing.T) {
	configureTestAuth()
	store := NewMemoryTokenStore()
	SetTokenStore(store)
	t.Cleanup(func() {
		SetTokenStore(nil)
	})

	pair, err := GenTokenPair(context.Background(), Claims{UserID: "user-1", Role: COMMON_USER})
	if err != nil {
		t.Fatalf("GenTokenPair() error = %v", err)
	}
	if pair.AToken == "" || pair.RToken == "" {
		t.Fatalf("GenTokenPair() = %+v, want non-empty atoken and rtoken", pair)
	}
	if pair.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", pair.TokenType)
	}
	if pair.ExpiresIn != config.Conf.Auth.ATokenExpire {
		t.Fatalf("ExpiresIn = %d, want %d", pair.ExpiresIn, config.Conf.Auth.ATokenExpire)
	}
	if pair.RefreshExpiresIn != config.Conf.Auth.RTokenExpire {
		t.Fatalf("RefreshExpiresIn = %d, want %d", pair.RefreshExpiresIn, config.Conf.Auth.RTokenExpire)
	}

	atokenClaims, err := ParseTokenString(pair.AToken)
	if err != nil {
		t.Fatalf("ParseTokenString(atoken) error = %v", err)
	}
	if atokenClaims.UserID != "user-1" || atokenClaims.Role != COMMON_USER || atokenClaims.TokenType != TokenTypeAToken {
		t.Fatalf("atoken claims = %+v", atokenClaims)
	}

	rtokenClaims, err := ParseRTokenString(pair.RToken)
	if err != nil {
		t.Fatalf("ParseRTokenString(rtoken) error = %v", err)
	}
	if rtokenClaims.UserID != "user-1" || rtokenClaims.Role != COMMON_USER || rtokenClaims.TokenType != TokenTypeRToken {
		t.Fatalf("rtoken claims = %+v", rtokenClaims)
	}

	if _, err := ParseTokenString(pair.RToken); !errors.Is(err, ErrTokenType) {
		t.Fatalf("ParseTokenString(rtoken) error = %v, want %v", err, ErrTokenType)
	}
	if _, err := ParseRTokenString(pair.AToken); !errors.Is(err, ErrTokenType) {
		t.Fatalf("ParseRTokenString(atoken) error = %v, want %v", err, ErrTokenType)
	}
}

func TestATokenBlacklistRevokesAccessToken(t *testing.T) {
	configureTestAuth()
	store := NewMemoryTokenStore()
	SetTokenStore(store)
	t.Cleanup(func() {
		SetTokenStore(nil)
	})

	pair, err := GenTokenPair(context.Background(), Claims{UserID: "user-1", Role: COMMON_USER})
	if err != nil {
		t.Fatalf("GenTokenPair() error = %v", err)
	}

	if err := BlacklistAToken(context.Background(), pair.AToken); err != nil {
		t.Fatalf("BlacklistAToken() error = %v", err)
	}
	if _, err := ParseTokenString(pair.AToken); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("ParseTokenString(revoked atoken) error = %v, want %v", err, ErrTokenRevoked)
	}
}

func TestRTokenRotationMakesOldRefreshTokenReplay(t *testing.T) {
	configureTestAuth()
	store := NewMemoryTokenStore()
	SetTokenStore(store)
	t.Cleanup(func() {
		SetTokenStore(nil)
	})

	pair, err := GenTokenPair(context.Background(), Claims{UserID: "user-1", Role: COMMON_USER})
	if err != nil {
		t.Fatalf("GenTokenPair() error = %v", err)
	}

	rotated, err := RotateRToken(context.Background(), pair.RToken)
	if err != nil {
		t.Fatalf("RotateRToken() error = %v", err)
	}
	if rotated.AToken == "" || rotated.RToken == "" || rotated.RToken == pair.RToken {
		t.Fatalf("RotateRToken() = %+v, want new token pair", rotated)
	}
	if _, err := ParseTokenString(rotated.AToken); err != nil {
		t.Fatalf("ParseTokenString(rotated atoken) error = %v", err)
	}
	if _, err := RotateRToken(context.Background(), pair.RToken); !errors.Is(err, ErrRTokenReplay) {
		t.Fatalf("RotateRToken(old rtoken) error = %v, want %v", err, ErrRTokenReplay)
	}
}

func TestRevokeTokenPairRevokesAccessAndRefreshTokens(t *testing.T) {
	configureTestAuth()
	store := NewMemoryTokenStore()
	SetTokenStore(store)
	t.Cleanup(func() {
		SetTokenStore(nil)
	})

	pair, err := GenTokenPair(context.Background(), Claims{UserID: "user-1", Role: COMMON_USER})
	if err != nil {
		t.Fatalf("GenTokenPair() error = %v", err)
	}

	if err := RevokeTokenPair(context.Background(), pair.AToken, pair.RToken); err != nil {
		t.Fatalf("RevokeTokenPair() error = %v", err)
	}
	if _, err := ParseTokenString(pair.AToken); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("ParseTokenString(revoked atoken) error = %v, want %v", err, ErrTokenRevoked)
	}
	if _, err := RotateRToken(context.Background(), pair.RToken); !errors.Is(err, ErrRTokenReplay) {
		t.Fatalf("RotateRToken(revoked rtoken) error = %v, want %v", err, ErrRTokenReplay)
	}
}

func configureTestAuth() {
	config.Conf.Auth = config.Auth{
		ATokenSecret: "test-atoken-secret",
		ATokenExpire: 60,
		RTokenSecret: "test-rtoken-secret",
		RTokenExpire: 300,
	}
}
