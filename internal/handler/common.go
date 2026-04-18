package handler

import (
	"encoding/json"
	"nurture/internal/constant"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type CommonHandler struct {
	commonLogic *logic.CommonLogic
}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{
		commonLogic: logic.NewCommonLogic(),
	}
}

func (h *CommonHandler) UploadFile(c *gin.Context) {
	global.Log.Infof("%s: %s", jwtx.GetUserID(c), c.Request.FormValue("file"))
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Response(c, nil, ErrFileRead)
		return
	}
	defer file.Close()

	url, err := h.commonLogic.UploadFile(c.Request.Context(), file, header)
	if err != nil {
		response.Response(c, nil, err)
		return
	}
	response.Response(c, url, nil)
}

// ChatStream AI 对话（SSE 流式响应）
func (h *CommonHandler) ChatStream(c *gin.Context) {
	req := middleware.GetBind[dto.ChatStreamReq](c)
	userID := jwtx.GetUserID(c)
	global.Log.Infof("%s: %v", userID, req)

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
	streamFunc := func(event dto.SSEEvent) {
		data, _ := json.Marshal(event)
		c.SSEvent("message", string(data))
		c.Writer.Flush()
	}

	// 执行对话
	_ = h.commonLogic.ChatStream(c.Request.Context(), userID, req, streamFunc)
}

// UploadKnowledge 上传知识库
func (h *CommonHandler) UploadKnowledge(c *gin.Context) {
	req := middleware.GetBind[dto.KnowledgeUploadReq](c)
	userID, role := jwtx.GetUserID(c), jwtx.GetRole(c)

	if req.SpaceType == constant.SPACE_TYPE_PUBLIC && role < jwtx.INTERNAL_USER {
		global.Log.Infof("%s: %d", userID, role)
		response.Response(c, nil, ErrPermissionDenied)
		return
	}

	global.Log.Infof("%s: %v", userID, req)
	err := h.commonLogic.UploadKnowledge(c.Request.Context(), userID, req)
	response.Response(c, nil, err)
}

// GetChatHistory 获取对话历史
func (h *CommonHandler) GetChatHistory(c *gin.Context) {
	req := middleware.GetBind[dto.ChatHistoryReq](c)
	userID := jwtx.GetUserID(c)
	global.Log.Infof("%s: %v", userID, req)
	resp, err := h.commonLogic.GetChatHistory(c.Request.Context(), userID, req)
	response.Response(c, resp, err)
}

// GrowthAnalysis 成长曲线分析
func (h *CommonHandler) GrowthAnalysis(c *gin.Context) {
	req := middleware.GetBind[dto.GrowthAnalysisReq](c)
	userID := jwtx.GetUserID(c)
	global.Log.Infof("GrowthAnalysis %s: %v", userID, req)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 流式回调
	streamFunc := func(event dto.SSEEvent) {
		data, _ := json.Marshal(event)
		c.SSEvent("message", string(data))
		c.Writer.Flush()
	}

	// 执行分析
	_ = h.commonLogic.GrowthAnalysisStream(c.Request.Context(), userID, req, streamFunc)
}

func (h *CommonHandler) GrowthReport(c *gin.Context) {
	req := middleware.GetBind[dto.GrowthReportReq](c)
	userID := jwtx.GetUserID(c)
	global.Log.Infof("GrowthReport %s: %v", userID, req)
	resp, err := h.commonLogic.GrowthReport(c.Request.Context(), userID, req)
	response.Response(c, resp, err)
}
