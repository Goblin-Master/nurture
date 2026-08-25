package file

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/upload", middleware.Authentication(jwtx.COMMON_USER), m.handler.Upload)
}
