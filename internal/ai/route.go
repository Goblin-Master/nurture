package ai

import (
	"nurture/internal/ai/dto"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	aiHandler := m.handler
	rg.POST("/knowledge/upload",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindJsonMiddleware[dto.KnowledgeUploadReq],
		aiHandler.UploadKnowledge,
	)
	rg.POST("/chat/stream",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindJsonMiddleware[dto.ChatStreamReq],
		aiHandler.ChatStream,
	)
	rg.GET("/chat/history",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindQueryMiddleware[dto.ChatHistoryReq],
		aiHandler.GetChatHistory,
	)
	rg.POST("/growth/analysis",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindJsonMiddleware[dto.GrowthAnalysisReq],
		aiHandler.GrowthAnalysis,
	)
	rg.POST("/report/growth",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindJsonMiddleware[dto.GrowthReportReq],
		aiHandler.GrowthReport,
	)
}
