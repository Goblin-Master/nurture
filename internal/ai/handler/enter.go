package handler

import (
	"encoding/json"
	aiconstant "nurture/internal/ai/constant"
	aidto "nurture/internal/ai/dto"
	ailogic "nurture/internal/ai/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AIHandler struct {
	aiLogic ailogic.IAILogic
	log     *zap.SugaredLogger
}

func NewAIHandler(aiLogic ailogic.IAILogic, log *zap.SugaredLogger) *AIHandler {
	return &AIHandler{
		aiLogic: aiLogic,
		log:     log,
	}
}

// ChatStream AI 对话（SSE 流式响应）
func (h *AIHandler) ChatStream(c *gin.Context) {
	req := middleware.GetBind[aidto.ChatStreamReq](c)
	userID := jwtx.GetUserID(c)
	if h.log != nil {
		h.log.Infof("%s: %v", userID, req)
	}

	// 如果没有 SessionID，返回错误
	if req.SessionID == "" {
		response.Response(c, nil, ErrSessionIDEmpty)
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 流式回调
	streamFunc := func(event aidto.SSEEvent) {
		data, _ := json.Marshal(event)
		c.SSEvent("message", string(data))
		c.Writer.Flush()
	}

	// 执行对话
	_ = h.aiLogic.ChatStream(c.Request.Context(), userID, req, streamFunc)
}

// UploadKnowledge 上传知识库
func (h *AIHandler) UploadKnowledge(c *gin.Context) {
	req := middleware.GetBind[aidto.KnowledgeUploadReq](c)
	userID, role := jwtx.GetUserID(c), jwtx.GetRole(c)

	if req.SpaceType == aiconstant.SpaceTypePublic && role < jwtx.INTERNAL_USER {
		if h.log != nil {
			h.log.Infof("%s: %d", userID, role)
		}
		response.Response(c, nil, ErrPermissionDenied)
		return
	}

	if h.log != nil {
		h.log.Infof("%s: %v", userID, req)
	}
	err := h.aiLogic.UploadKnowledge(c.Request.Context(), userID, req)
	response.Response(c, nil, err)
}

// GetChatHistory 获取对话历史
func (h *AIHandler) GetChatHistory(c *gin.Context) {
	req := middleware.GetBind[aidto.ChatHistoryReq](c)
	userID := jwtx.GetUserID(c)
	if h.log != nil {
		h.log.Infof("%s: %v", userID, req)
	}
	resp, err := h.aiLogic.GetChatHistory(c.Request.Context(), userID, req)
	response.Response(c, resp, err)
}

// GrowthAnalysis 成长曲线分析
func (h *AIHandler) GrowthAnalysis(c *gin.Context) {
	req := middleware.GetBind[aidto.GrowthAnalysisReq](c)
	userID := jwtx.GetUserID(c)
	if h.log != nil {
		h.log.Infof("GrowthAnalysis %s: %v", userID, req)
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 流式回调
	streamFunc := func(event aidto.SSEEvent) {
		data, _ := json.Marshal(event)
		c.SSEvent("message", string(data))
		c.Writer.Flush()
	}

	// 执行分析
	_ = h.aiLogic.GrowthAnalysisStream(c.Request.Context(), userID, req, streamFunc)
}

func (h *AIHandler) GrowthReport(c *gin.Context) {
	req := middleware.GetBind[aidto.GrowthReportReq](c)
	userID := jwtx.GetUserID(c)
	if h.log != nil {
		h.log.Infof("GrowthReport %s: %v", userID, req)
	}
	resp, err := h.aiLogic.GrowthReport(c.Request.Context(), userID, req)
	response.Response(c, resp, err)
}
