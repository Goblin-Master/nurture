package middleware

import (
	"nurture/internal/config"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:  []string{"Authorization", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		//是否允许你带cookie之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if config.Conf.App.Env == "dev" {
				return true
			}
			if strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "https://127.0.0.1") || strings.HasPrefix(origin, "https://localhost") {
				return true
			}
			if strings.HasPrefix(origin, "http://10") || strings.HasPrefix(origin, "https://10") ||
				strings.HasPrefix(origin, "http://192.168.") || strings.HasPrefix(origin, "https://192.168.") {
				return true
			}
			return false
		},
		MaxAge: 12 * time.Hour,
	})
}
