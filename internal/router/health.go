package router

import (
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		response.Response(c, "ok", nil)
	})
}
