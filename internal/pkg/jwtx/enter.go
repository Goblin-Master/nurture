package jwtx

import (
	"errors"
	"fmt"
	"nurture/internal/config"
	"nurture/internal/constant"
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

type MyClaims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
}

var (
	ErrDefault          = errors.New("jwt default error")
	ErrTokenEmpty       = errors.New("token is empty")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrPermissionDenied = errors.New("permission denied")
)

func GenToken(c Claims) (string, error) {
	secret := config.Conf.Auth.AccessSecret
	expiredTime := config.Conf.Auth.AccessExpire
	// 创建一个我们自己的声明
	claims := MyClaims{
		c.UserID,
		c.Role,
		jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiredTime) * time.Second)), // 过期时间
			Issuer:    "Nurture",
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString([]byte(secret))
}

// GenTestToken 用于测试生成 Token，需要确保 config 已加载
func GenTestToken(userID string, role Role) (string, error) {
	return GenToken(Claims{
		UserID: userID,
		Role:   role,
	})
}

func ParseToken(c *gin.Context) (string, Role, error) {
	data := c.GetHeader("Authorization")
	if data == "" {
		return "", 0, ErrTokenEmpty
	}
	token := strings.TrimPrefix(data, "Bearer ")
	claims, err := ParseTokenString(token)
	if err != nil {
		return "", 0, err
	}
	return claims.UserID, claims.Role, nil
}

func ParseTokenString(token string) (*MyClaims, error) {
	// 解析token
	var claims MyClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Conf.Auth.AccessSecret), nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, ErrTokenExpired
		}
		if strings.Contains(err.Error(), "signature is invalid") {
			return nil, ErrTokenInvalid
		}
		if strings.Contains(err.Error(), "token contains an invalid") {
			return nil, ErrTokenInvalid
		}
		fmt.Println(err)
		return nil, ErrDefault
	}
	if claims, ok := t.Claims.(*MyClaims); ok && t.Valid {
		return claims, nil
	}
	return nil, ErrDefault
}

// 必须使用了鉴权中间件才能用
func GetUserID(c *gin.Context) string {
	if data, exists := c.Get(constant.TOKEN_USER_ID); exists {
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
	if data, exists := c.Get(constant.TOKEN_ROLE); exists {
		role, ok := data.(Role)
		if ok {
			return role
		}
	}
	return 0
}
