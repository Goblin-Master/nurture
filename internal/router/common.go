package router

import (
	"nurture/internal/dto"
	"nurture/internal/handler"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) {
		response.Response(c, "ok", nil)
	})
}

func registerCommonRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		response.Response(c, "pong", nil)
	})
	commonHandler := handler.NewCommonHandler()
	rg.POST("/file/upload", middleware.Authentication(jwtx.COMMON_USER), commonHandler.UploadFile)

	ai := rg.Group("/ai")
	{
		ai.POST("/knowledge/upload",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.KnowledgeUploadReq],
			commonHandler.UploadKnowledge,
		)

		ai.POST("/chat/stream",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.ChatStreamReq],
			commonHandler.ChatStream,
		)

		ai.GET("/chat/history",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.ChatHistoryReq],
			commonHandler.GetChatHistory,
		)

		ai.POST("/growth/analysis",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.GrowthAnalysisReq],
			commonHandler.GrowthAnalysis,
		)

		ai.POST("/report/growth",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.GrowthReportReq],
			commonHandler.GrowthReport,
		)
	}
}
