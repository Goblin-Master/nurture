package middleware

import (
	"net/http"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func RequireDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.Conf.DB.Enable || global.DB == nil {
			c.JSON(http.StatusServiceUnavailable, response.Body{
				Code:    -1,
				Message: "数据库服务未启用",
				Data:    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
