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

// 成长分析
type (
	GrowthAnalysisReq struct {
		Birthday int64        `json:"birthday" binding:"required"` // 13位毫秒级Unix时间戳
		Metric   string       `json:"metric" binding:"required,oneof=height weight head_circumference"`
		Unit     string       `json:"unit" binding:"required"` // cm / kg
		Items    []GrowthItem `json:"items" binding:"required,dive"`
	}
	GrowthItem struct {
		Time  int64   `json:"time" binding:"required"` // 13位毫秒级Unix时间戳
		Value float64 `json:"value" binding:"required"`
	}
)

type (
	GrowthReportReq struct {
		BabyID    string `json:"baby_id" binding:"required"`
		RangeDays int    `json:"range_days"`
		Language  string `json:"language"`
	}
	GrowthReportResp struct {
		Markdown string           `json:"markdown"`
		Data     GrowthReportData `json:"data"`
	}
)

type (
	GrowthReportData struct {
		Baby     GrowthReportBaby     `json:"baby"`
		Range    GrowthReportRange    `json:"range"`
		Growth   GrowthReportGrowth   `json:"growth"`
		Analysis GrowthReportAnalysis `json:"analysis"`
	}
	GrowthReportBaby struct {
		BabyID   string `json:"baby_id"`
		Name     string `json:"name"`
		Gender   string `json:"gender"`
		Birthday int64  `json:"birthday"`
		Avatar   string `json:"avatar"`
	}
	GrowthReportRange struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
		Days int   `json:"days"`
	}
	GrowthReportGrowth struct {
		Items []GrowthReportGrowthItem `json:"items"`
	}
	GrowthReportGrowthItem struct {
		Time              int64    `json:"time"`
		Height            *float64 `json:"height,omitempty"`
		Weight            *float64 `json:"weight,omitempty"`
		HeadCircumference *float64 `json:"head_circumference,omitempty"`
		Remark            string   `json:"remark,omitempty"`
	}
)

type (
	GrowthReportAnalysis struct {
		Height            GrowthMetricAnalysis `json:"height"`
		Weight            GrowthMetricAnalysis `json:"weight"`
		HeadCircumference GrowthMetricAnalysis `json:"head_circumference"`
	}
	GrowthMetricAnalysis struct {
		Points    int      `json:"points"`
		FirstTime int64    `json:"first_time,omitempty"`
		LastTime  int64    `json:"last_time,omitempty"`
		First     *float64 `json:"first,omitempty"`
		Last      *float64 `json:"last,omitempty"`
		Delta     *float64 `json:"delta,omitempty"`
		PerWeek   *float64 `json:"per_week,omitempty"`
		PerMonth  *float64 `json:"per_month,omitempty"`
		Predict30 *float64 `json:"predict_30,omitempty"`
	}
)
