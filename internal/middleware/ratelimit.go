package middleware

import (
	"fmt"
	"nurture/internal/global"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/ratelimitx"
	"nurture/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

var rl = ratelimitx.NewWindowLimiter(nil)

func RateLimitUser(key string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.RDB != nil {
			rl.SetRedis(global.RDB)
		}
		userID := jwtx.GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}
		ok, _, _ := rl.Allow(c.Request.Context(), fmt.Sprintf("rl:%s:%s", key, userID), limit, window)
		if !ok {
			c.JSON(429, response.Body{Code: -1, Message: "请求过于频繁", Data: nil})
			c.Abort()
			return
		}
		c.Next()
	}
}
