package router

import (
	"nurture/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerWSRoutes(rg *gin.RouterGroup) {
	wsHandler := handler.NewWebSocketHandler()
	rg.GET("/chat", wsHandler.Connect)
	rg.GET("/groups", wsHandler.ConnectGroups)
}
