package dto

// 知识库上传
type (
	KnowledgeUploadReq struct {
		SpaceType string `json:"space_type" binding:"required,oneof=private public"`
		Content   string `json:"content" binding:"required"`
	}
	KnowledgeUploadResp struct {
		Message string `json:"message"`
	}
)

// 知识库配置
type KBConfig struct {
	Enable        bool `json:"enable" default:"false"`
	SearchPrivate bool `json:"search_private" default:"false"`
	SearchPublic  bool `json:"search_public" default:"false"`
	TopK          int  `json:"top_k" default:"3"`
}

// AI 对话
type (
	ChatStreamReq struct {
		SessionID string   `json:"session_id" binding:"required"`
		Message   string   `json:"message" binding:"required"`
		Images    []string `json:"images"`
		KBConfig  KBConfig `json:"kb_config"`
	}
)

// SSE 响应事件
type SSEEvent struct {
	Type      string `json:"type"` // content / error / done
	Content   string `json:"content,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// 获取对话历史
type (
	ChatHistoryReq struct {
		SessionID string `form:"session_id" binding:"required"`
	}
	ChatHistoryResp struct {
		Messages []ChatMessageItem `json:"messages"`
	}
	ChatMessageItem struct {
		Role      string   `json:"role"`
		Content   string   `json:"content"`
		Images    []string `json:"images,omitempty"`
		Timestamp int64    `json:"timestamp"`
	}
)
