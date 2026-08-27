package jwtx

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"nurture/internal/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Role int

const (
	COMMON_USER = iota + 1
	INTERNAL_USER
	ADMIN
)

const (
	ContextUserIDKey = "UserID"
	ContextRoleKey   = "Role"
)

const (
	TokenTypeAToken = "atoken"
	TokenTypeRToken = "rtoken"
)

type MyClaims struct {
	UserID    string `json:"user_id"`
	Role      Role   `json:"role"`
	TokenType string `json:"token_type"`
	Nonce     string `json:"nonce,omitempty"`
	jwt.RegisteredClaims
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
}

type TokenPair struct {
	AToken           string `json:"atoken"`
	RToken           string `json:"rtoken"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

var (
	ErrDefault          = errors.New("jwt default error")
	ErrTokenEmpty       = errors.New("token is empty")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenType        = errors.New("token type is invalid")
	ErrTokenRevoked     = errors.New("token has been revoked")
	ErrTokenStore       = errors.New("token store unavailable")
	ErrRTokenReplay     = errors.New("refresh token has been used")
	ErrPermissionDenied = errors.New("permission denied")
)

var tokenStore TokenStore

func SetTokenStore(store TokenStore) {
	tokenStore = store
}

func GenToken(c Claims) (string, error) {
	return genToken(c, TokenTypeAToken, config.Conf.Auth.ATokenSecret, authDuration(config.Conf.Auth.ATokenExpire), time.Now())
}

// GenTestToken 用于测试生成 Token，需要确保 config 已加载
func GenTestToken(userID string, role Role) (string, error) {
	return GenToken(Claims{
		UserID: userID,
		Role:   role,
	})
}

func GenTokenPair(ctx context.Context, c Claims) (TokenPair, error) {
	store := tokenStore
	if store == nil {
		return TokenPair{}, ErrTokenStore
	}
	pair, err := genTokenPair(c)
	if err != nil {
		return TokenPair{}, err
	}
	err = store.SetActiveRToken(ctx, TokenHash(pair.RToken), RTokenSession{
		UserID: c.UserID,
		Role:   c.Role,
	}, authDuration(config.Conf.Auth.RTokenExpire))
	if err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func RotateRToken(ctx context.Context, rtoken string) (TokenPair, error) {
	store := tokenStore
	if store == nil {
		return TokenPair{}, ErrTokenStore
	}
	claims, err := parseTokenString(ctx, rtoken, config.Conf.Auth.RTokenSecret, TokenTypeRToken, false)
	if err != nil {
		return TokenPair{}, err
	}
	usedTTL := tokenTTL(claims)
	if usedTTL <= 0 {
		return TokenPair{}, ErrTokenExpired
	}
	pair, err := genTokenPair(Claims{UserID: claims.UserID, Role: claims.Role})
	if err != nil {
		return TokenPair{}, err
	}
	err = store.RotateRToken(
		ctx,
		TokenHash(rtoken),
		TokenHash(pair.RToken),
		RTokenSession{UserID: claims.UserID, Role: claims.Role},
		authDuration(config.Conf.Auth.RTokenExpire),
		usedTTL,
	)
	if err != nil {
		return TokenPair{}, err
	}
	return pair, nil
}

func RevokeTokenPair(ctx context.Context, atoken, rtoken string) error {
	store := tokenStore
	if store == nil {
		return ErrTokenStore
	}
	if err := BlacklistAToken(ctx, atoken); err != nil && !errors.Is(err, ErrTokenExpired) {
		return err
	}
	claims, err := parseTokenString(ctx, rtoken, config.Conf.Auth.RTokenSecret, TokenTypeRToken, false)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil
		}
		return err
	}
	ttl := tokenTTL(claims)
	hash := TokenHash(rtoken)
	if err := store.DeleteActiveRToken(ctx, hash); err != nil {
		return err
	}
	if ttl > 0 {
		return store.MarkRTokenUsed(ctx, hash, ttl)
	}
	return nil
}

func BlacklistAToken(ctx context.Context, atoken string) error {
	store := tokenStore
	if store == nil {
		return ErrTokenStore
	}
	claims, err := parseTokenString(ctx, atoken, config.Conf.Auth.ATokenSecret, TokenTypeAToken, false)
	if err != nil {
		return err
	}
	ttl := tokenTTL(claims)
	if ttl <= 0 {
		return ErrTokenExpired
	}
	return store.BlacklistAToken(ctx, TokenHash(atoken), ttl)
}

func ParseToken(c *gin.Context) (string, Role, error) {
	token, err := BearerToken(c)
	if err != nil {
		return "", 0, err
	}
	claims, err := parseTokenString(c.Request.Context(), token, config.Conf.Auth.ATokenSecret, TokenTypeAToken, true)
	if err != nil {
		return "", 0, err
	}
	return claims.UserID, claims.Role, nil
}

func ParseTokenString(token string) (*MyClaims, error) {
	return parseTokenString(context.Background(), token, config.Conf.Auth.ATokenSecret, TokenTypeAToken, true)
}

func ParseRTokenString(token string) (*MyClaims, error) {
	return parseTokenString(context.Background(), token, config.Conf.Auth.RTokenSecret, TokenTypeRToken, false)
}

// 必须使用了鉴权中间件才能用
func GetUserID(c *gin.Context) string {
	if data, exists := c.Get(ContextUserIDKey); exists {
		user_id, ok := data.(string)
		if ok {
			return user_id
		}
	}
	auth := c.GetHeader("Authorization")
	if auth != "" {
		token := strings.TrimPrefix(auth, "Bearer ")
		if claims, err := ParseTokenString(token); err == nil {
			return claims.UserID
		}
	}
	return ""
}

// 必须使用了鉴权中间件才能用
func GetRole(c *gin.Context) Role {
	if data, exists := c.Get(ContextRoleKey); exists {
		role, ok := data.(Role)
		if ok {
			return role
		}
	}
	return 0
}

func BearerToken(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return "", ErrTokenEmpty
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	if token == "" {
		return "", ErrTokenEmpty
	}
	return token, nil
}

func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func genTokenPair(c Claims) (TokenPair, error) {
	now := time.Now()
	atoken, err := genToken(c, TokenTypeAToken, config.Conf.Auth.ATokenSecret, authDuration(config.Conf.Auth.ATokenExpire), now)
	if err != nil {
		return TokenPair{}, err
	}
	rtoken, err := genToken(c, TokenTypeRToken, config.Conf.Auth.RTokenSecret, authDuration(config.Conf.Auth.RTokenExpire), now)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AToken:           atoken,
		RToken:           rtoken,
		TokenType:        "Bearer",
		ExpiresIn:        config.Conf.Auth.ATokenExpire,
		RefreshExpiresIn: config.Conf.Auth.RTokenExpire,
	}, nil
}

func genToken(c Claims, tokenType, secret string, ttl time.Duration, now time.Time) (string, error) {
	if strings.TrimSpace(secret) == "" || ttl <= 0 {
		return "", ErrDefault
	}
	nonce, err := randomNonce()
	if err != nil {
		return "", err
	}
	claims := MyClaims{
		UserID:    c.UserID,
		Role:      c.Role,
		TokenType: tokenType,
		Nonce:     nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "Nurture",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func parseTokenString(ctx context.Context, token, secret, wantType string, checkBlacklist bool) (*MyClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrTokenEmpty
	}
	var unverified MyClaims
	if _, _, err := new(jwt.Parser).ParseUnverified(token, &unverified); err == nil && unverified.TokenType != "" && unverified.TokenType != wantType {
		return nil, ErrTokenType
	}
	var claims MyClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			}
			return nil, ErrTokenInvalid
		}
		return nil, ErrDefault
	}
	parsedClaims, ok := t.Claims.(*MyClaims)
	if !ok || !t.Valid {
		return nil, ErrTokenInvalid
	}
	if parsedClaims.TokenType != wantType {
		return nil, ErrTokenType
	}
	if checkBlacklist && tokenStore != nil {
		revoked, err := tokenStore.IsATokenBlacklisted(ctx, TokenHash(token))
		if err != nil {
			return nil, err
		}
		if revoked {
			return nil, ErrTokenRevoked
		}
	}
	return parsedClaims, nil
}

func authDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}

func tokenTTL(claims *MyClaims) time.Duration {
	if claims == nil || claims.ExpiresAt == nil {
		return 0
	}
	return time.Until(claims.ExpiresAt.Time)
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
