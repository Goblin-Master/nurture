package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	aiconstant "nurture/internal/ai/constant"
	aidto "nurture/internal/ai/dto"
	aihandler "nurture/internal/ai/handler"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func TestAIHandlerChatStreamWritesReturnedErrorEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := aihandler.NewAIHandler(&aiLogicFake{chatErr: errors.New("stream failed")}, nil)
	w := performChatStreamRequest(t, h, aidto.ChatStreamReq{
		SessionID: "session-1",
		Message:   "hello",
	})

	body := w.Body.String()
	if !strings.Contains(body, `"type":"error"`) || !strings.Contains(body, `"error":"stream failed"`) {
		t.Fatalf("ChatStream body = %q, want returned error SSE event", body)
	}
}

func TestAIHandlerChatStreamDoesNotDuplicateLogicErrorEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := aihandler.NewAIHandler(&aiLogicFake{
		chatErr: errors.New("stream failed"),
		chatEvent: aidto.SSEEvent{
			Type:  aiconstant.SSETypeError,
			Error: "already streamed",
		},
	}, nil)
	w := performChatStreamRequest(t, h, aidto.ChatStreamReq{
		SessionID: "session-1",
		Message:   "hello",
	})

	body := w.Body.String()
	if strings.Count(body, `"type":"error"`) != 1 {
		t.Fatalf("ChatStream body = %q, want one error event", body)
	}
	if !strings.Contains(body, `"error":"already streamed"`) {
		t.Fatalf("ChatStream body = %q, want logic error event", body)
	}
}

func TestAIHandlerGrowthAnalysisWritesReturnedErrorEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := aihandler.NewAIHandler(&aiLogicFake{growthAnalysisErr: errors.New("growth failed")}, nil)
	w := performGrowthAnalysisRequest(t, h, aidto.GrowthAnalysisReq{
		Birthday: 1704067200000,
		Metric:   "height",
		Unit:     "cm",
		Items: []aidto.GrowthItem{
			{Time: 1704067200000, Value: 50},
		},
	})

	body := w.Body.String()
	if !strings.Contains(body, `"type":"error"`) || !strings.Contains(body, `"error":"growth failed"`) {
		t.Fatalf("GrowthAnalysis body = %q, want returned error SSE event", body)
	}
}

func performChatStreamRequest(t *testing.T, h *aihandler.AIHandler, req aidto.ChatStreamReq) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	r.POST("/api/ai/chat/stream",
		injectUser,
		middleware.BindJsonMiddleware[aidto.ChatStreamReq],
		h.ChatStream,
	)
	return performJSONRequest(t, r, "/api/ai/chat/stream", req)
}

func performGrowthAnalysisRequest(t *testing.T, h *aihandler.AIHandler, req aidto.GrowthAnalysisReq) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	r.POST("/api/ai/growth/analysis",
		injectUser,
		middleware.BindJsonMiddleware[aidto.GrowthAnalysisReq],
		h.GrowthAnalysis,
	)
	return performJSONRequest(t, r, "/api/ai/growth/analysis", req)
}

func performJSONRequest(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal request failed: %v", err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func injectUser(c *gin.Context) {
	c.Set(jwtx.ContextUserIDKey, "user-1")
	c.Next()
}

type aiLogicFake struct {
	chatErr           error
	chatEvent         aidto.SSEEvent
	growthAnalysisErr error
}

func (f *aiLogicFake) ChatStream(ctx context.Context, userID string, req aidto.ChatStreamReq, streamFunc func(event aidto.SSEEvent)) error {
	if f.chatEvent.Type != "" {
		streamFunc(f.chatEvent)
	}
	return f.chatErr
}

func (f *aiLogicFake) UploadKnowledge(context.Context, string, aidto.KnowledgeUploadReq) error {
	return nil
}

func (f *aiLogicFake) GetChatHistory(context.Context, string, aidto.ChatHistoryReq) (aidto.ChatHistoryResp, error) {
	return aidto.ChatHistoryResp{}, nil
}

func (f *aiLogicFake) GrowthAnalysisStream(ctx context.Context, userID string, req aidto.GrowthAnalysisReq, streamFunc func(event aidto.SSEEvent)) error {
	return f.growthAnalysisErr
}

func (f *aiLogicFake) GrowthReport(context.Context, string, aidto.GrowthReportReq) (aidto.GrowthReportResp, error) {
	return aidto.GrowthReportResp{}, nil
}
