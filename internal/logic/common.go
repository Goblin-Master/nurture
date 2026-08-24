package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/pkg/aix"
	"nurture/internal/repo"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/tmc/langchaingo/llms"
)

type ICommonLogic interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	ChatStream(ctx context.Context, userID string, req dto.ChatStreamReq, streamFunc func(event dto.SSEEvent)) error
	UploadKnowledge(ctx context.Context, userID string, req dto.KnowledgeUploadReq) error
	GetChatHistory(ctx context.Context, userID string, req dto.ChatHistoryReq) (dto.ChatHistoryResp, error)
	GrowthAnalysisStream(ctx context.Context, userID string, req dto.GrowthAnalysisReq, streamFunc func(event dto.SSEEvent)) error
	GrowthReport(ctx context.Context, userID string, req dto.GrowthReportReq) (dto.GrowthReportResp, error)
}

type CommonLogic struct {
	aiRepo   *repo.AIRepo
	babyRepo *repo.BabyRepo
}

func NewCommonLogic() *CommonLogic {
	return &CommonLogic{
		aiRepo:   repo.NewAIRepo(),
		babyRepo: repo.NewBabyRepo(),
	}
}

var _ ICommonLogic = (*CommonLogic)(nil)

func aiAvailable() bool {
	return config.Conf.AI.Enable && global.AIX != nil
}

func streamAIUnavailable(streamFunc func(event dto.SSEEvent)) {
	if streamFunc == nil {
		return
	}
	streamFunc(dto.SSEEvent{
		Type:  constant.SSE_TYPE_ERROR,
		Error: ErrChatStream.Error(),
	})
}

func (l *CommonLogic) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if !config.Conf.Minio.Enable || global.MIO == nil {
		return "", ErrFileUpload
	}

	// 1. Calculate MD5
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		global.Log.Error(err)
		return "", ErrFileRead
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		global.Log.Error(err)
		return "", ErrFileRead
	}

	// 2. Generate filename
	ext := filepath.Ext(header.Filename)
	objectName := fileHash + ext

	// 3. Construct URL
	protocol := "http"
	if config.Conf.Minio.UseSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, config.Conf.Minio.Endpoint, config.Conf.Minio.Bucket, objectName)

	// 4. Check if file exists
	_, err := global.MIO.StatObject(ctx, config.Conf.Minio.Bucket, objectName, minio.StatObjectOptions{})
	if err == nil {
		return url, nil
	}

	// 5. Upload to MinIO
	_, err = global.MIO.PutObject(ctx, config.Conf.Minio.Bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		global.Log.Error(err)
		return "", ErrFileUpload
	}

	return url, nil
}

// ChatStream 流式对话
func (l *CommonLogic) ChatStream(ctx context.Context, userID string, req dto.ChatStreamReq,
	streamFunc func(event dto.SSEEvent)) error {
	if !aiAvailable() {
		streamAIUnavailable(streamFunc)
		return ErrChatStream
	}

	// 1. 获取最近 3 轮对话历史（6 条消息）作为 AI 上下文
	history, err := l.aiRepo.GetRecentHistory(ctx, userID, req.SessionID, constant.AI_CONTEXT_MESSAGES)
	if err != nil {
		global.Log.Error(err)
		// 历史获取失败不阻断，继续对话
		history = []aix.ChatMessage{}
	}

	var extraContext string
	if req.AutoContext && strings.TrimSpace(req.BabyID) != "" {
		days := req.ContextDays
		if days <= 0 || days > 180 {
			days = 30
		}
		b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
		if err != nil {
			if errors.Is(err, repo.ErrBabyNotExist) {
				return ErrBabyNotExist
			}
			global.Log.Error(err)
			return ErrDefault
		}
		to := time.Now().UnixMilli()
		from := to - int64(days)*24*60*60*1000
		rows, err := l.babyRepo.ListGrowthRecordsByBabyIDBetween(ctx, req.BabyID, from, to)
		if err != nil {
			global.Log.Error(err)
			return ErrDefault
		}
		items := make([]dto.GrowthReportGrowthItem, 0, len(rows))
		for _, r := range rows {
			it := dto.GrowthReportGrowthItem{Time: r.RecordTime}
			if r.Height.Valid {
				v := r.Height.Float64
				it.Height = &v
			}
			if r.Weight.Valid {
				v := r.Weight.Float64
				it.Weight = &v
			}
			if r.HeadCircumference.Valid {
				v := r.HeadCircumference.Float64
				it.HeadCircumference = &v
			}
			items = append(items, it)
		}
		totalPoints := len(items)
		truncated := false
		if len(items) > 60 {
			items = items[len(items)-60:]
			truncated = true
		}
		data := dto.GrowthReportData{
			Baby: dto.GrowthReportBaby{
				BabyID:   b.BabyID.String(),
				Name:     b.Name,
				Gender:   b.Gender,
				Birthday: b.Birthday,
				Avatar:   b.Avatar,
			},
			Range: dto.GrowthReportRange{
				From: from,
				To:   to,
				Days: days,
			},
			Growth: dto.GrowthReportGrowth{
				Items: items,
			},
			Analysis: dto.GrowthReportAnalysis{
				Height:            analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.Height }),
				Weight:            analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.Weight }),
				HeadCircumference: analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.HeadCircumference }),
			},
		}
		payload := struct {
			Data        dto.GrowthReportData `json:"data"`
			TotalPoints int                  `json:"total_points"`
			Truncated   bool                 `json:"truncated"`
		}{
			Data:        data,
			TotalPoints: totalPoints,
			Truncated:   truncated,
		}
		if bts, e := json.Marshal(payload); e == nil {
			extraContext = string(bts)
		} else {
			global.Log.Error(e)
		}
	}

	// 2. RAG 检索
	var ragContext string
	// 构建需要检索的知识库集合
	collections := l.buildCollections(userID)

	// 如果有选中的知识库，则进行检索
	if len(collections) > 0 {
		topK := config.Conf.AI.KBConfig.TopK
		if topK <= 0 {
			topK = config.Conf.AI.Retrieval.DefaultTopK
		}

		docs, e := l.aiRepo.SimilaritySearch(ctx, req.Message, collections, topK)
		if e != nil {
			// 检索失败记录日志，但不阻断对话，仅降级为普通对话
			global.Log.Errorf("RAG SimilaritySearch failed: %v", e)
		}
		if len(docs) > 0 {
			// 拼接检索结果
			var sb strings.Builder
			for i, doc := range docs {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc.PageContent))
			}
			ragContext = sb.String()
		}
	}

	// 3. 构建消息
	messages := global.AIX.BuildMessages(history, req.Message, req.Images, ragContext, extraContext)

	// 4. 流式对话
	fullResponse, err := global.AIX.StreamChat(ctx, messages, func(chunk string) {
		streamFunc(dto.SSEEvent{
			Type:    constant.SSE_TYPE_CONTENT,
			Content: chunk,
		})
	})
	if err != nil {
		global.Log.Error(err)
		streamFunc(dto.SSEEvent{
			Type:  constant.SSE_TYPE_ERROR,
			Error: ErrChatStream.Error(),
		})
		return ErrChatStream
	}

	// 5. 保存对话历史
	now := time.Now().Unix()
	_ = l.aiRepo.SaveMessage(ctx, userID, req.SessionID, aix.ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Images:    req.Images,
		Timestamp: now,
	})
	_ = l.aiRepo.SaveMessage(ctx, userID, req.SessionID, aix.ChatMessage{
		Role:      "assistant",
		Content:   fullResponse,
		Timestamp: now,
	})

	// 6. 发送完成事件
	streamFunc(dto.SSEEvent{
		Type:      constant.SSE_TYPE_DONE,
		SessionID: req.SessionID,
	})

	return nil
}

// UploadKnowledge 上传知识库
func (l *CommonLogic) UploadKnowledge(ctx context.Context, userID string, req dto.KnowledgeUploadReq) error {
	if !aiAvailable() {
		return ErrKnowledgeUpload
	}

	// 构建 CollectionName
	var collectionName string
	switch req.SpaceType {
	case constant.SPACE_TYPE_PRIVATE:
		collectionName = fmt.Sprintf(constant.COLLECTION_USER_PREFIX, userID)
	case constant.SPACE_TYPE_PUBLIC:
		collectionName = constant.COLLECTION_PUBLIC
	default:
		return ErrInvalidSpaceType
	}

	// 添加文档
	err := l.aiRepo.AddDocument(ctx, collectionName, req.Content)
	if err != nil {
		global.Log.Error(err)
		return ErrKnowledgeUpload
	}
	return nil
}

// GetChatHistory 获取完整对话历史（供前端展示）
func (l *CommonLogic) GetChatHistory(ctx context.Context, userID string, req dto.ChatHistoryReq) (dto.ChatHistoryResp, error) {
	var resp dto.ChatHistoryResp

	history, err := l.aiRepo.GetFullHistory(ctx, userID, req.SessionID)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}

	resp.Messages = make([]dto.ChatMessageItem, len(history))
	for i, msg := range history {
		resp.Messages[i] = dto.ChatMessageItem{
			Role:      msg.Role,
			Content:   msg.Content,
			Images:    msg.Images,
			Timestamp: msg.Timestamp,
		}
	}

	return resp, nil
}

func (l *CommonLogic) buildCollections(userID string) []string {
	var collections []string
	cfg := config.Conf.AI.KBConfig
	if !cfg.Enable {
		return collections
	}
	if cfg.SearchPrivate {
		collections = append(collections, fmt.Sprintf(constant.COLLECTION_USER_PREFIX, userID))
	}
	if cfg.SearchPublic {
		collections = append(collections, constant.COLLECTION_PUBLIC)
	}
	return collections
}

// GrowthAnalysisStream 成长曲线分析
func (l *CommonLogic) GrowthAnalysisStream(ctx context.Context, userID string, req dto.GrowthAnalysisReq, streamFunc func(event dto.SSEEvent)) error {
	// 1. 验证单位
	// 1. 验证单位
	switch req.Metric {
	case "height", "head_circumference":
		if req.Unit != "cm" {
			return fmt.Errorf("invalid unit for %s: expected cm, got %s", req.Metric, req.Unit)
		}
	case "weight":
		if req.Unit != "kg" {
			return fmt.Errorf("invalid unit for %s: expected kg, got %s", req.Metric, req.Unit)
		}
	}
	if !aiAvailable() {
		streamAIUnavailable(streamFunc)
		return ErrChatStream
	}

	// 2. 构建提示词
	var sb strings.Builder
	sb.WriteString("请作为专业的儿科医生，分析以下宝宝的生长发育数据，并给出评估意见和建议。\n\n")

	// 基础信息
	birthday := time.UnixMilli(req.Birthday).Format("2006-01-02")
	metricName := map[string]string{
		"height":             "身高",
		"weight":             "体重",
		"head_circumference": "头围",
	}[req.Metric]

	sb.WriteString(fmt.Sprintf("宝宝出生日期：%s\n", birthday))
	sb.WriteString(fmt.Sprintf("测量指标：%s (%s)\n", metricName, req.Unit))
	sb.WriteString("测量记录：\n")

	// 记录列表
	for _, item := range req.Items {
		recordTime := time.UnixMilli(item.Time).Format("2006-01-02")
		// 计算月龄
		ageDays := int(time.UnixMilli(item.Time).Sub(time.UnixMilli(req.Birthday)).Hours() / 24)
		ageMonths := float64(ageDays) / 30.0
		sb.WriteString(fmt.Sprintf("- %s (约%.1f个月): %.2f\n", recordTime, ageMonths, item.Value))
	}

	sb.WriteString("\n请分析：\n1. 生长趋势是否正常？\n2. 与标准曲线相比处于什么水平？\n3. 有什么具体的喂养或护理建议？")

	prompt := sb.String()

	// 3. 调用 AI
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一个专业的儿科医生助手，擅长分析婴幼儿生长发育数据。"),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	_, err := global.AIX.StreamChat(ctx, messages, func(chunk string) {
		streamFunc(dto.SSEEvent{
			Type:    constant.SSE_TYPE_CONTENT,
			Content: chunk,
		})
	})

	if err != nil {
		global.Log.Error(err)
		streamFunc(dto.SSEEvent{
			Type:  constant.SSE_TYPE_ERROR,
			Error: ErrChatStream.Error(),
		})
		return ErrChatStream
	}

	// 4. 完成
	streamFunc(dto.SSEEvent{
		Type: constant.SSE_TYPE_DONE,
	})

	return nil
}

func (l *CommonLogic) GrowthReport(ctx context.Context, userID string, req dto.GrowthReportReq) (dto.GrowthReportResp, error) {
	var resp dto.GrowthReportResp
	if strings.TrimSpace(req.BabyID) == "" {
		return resp, ErrParamsType
	}
	days := req.RangeDays
	if days <= 0 || days > 365 {
		days = 90
	}
	to := time.Now().UnixMilli()
	from := time.Now().AddDate(0, 0, -days).UnixMilli()

	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	rows, err := l.babyRepo.ListGrowthRecordsByBabyIDBetween(ctx, req.BabyID, from, to)
	if err != nil {
		if errors.Is(err, repo.ErrDefault) {
			global.Log.Error(err)
		}
		return resp, ErrDefault
	}

	items := make([]dto.GrowthReportGrowthItem, 0, len(rows))
	for _, r := range rows {
		var h *float64
		if r.Height.Valid {
			v := r.Height.Float64
			h = &v
		}
		var w *float64
		if r.Weight.Valid {
			v := r.Weight.Float64
			w = &v
		}
		var hc *float64
		if r.HeadCircumference.Valid {
			v := r.HeadCircumference.Float64
			hc = &v
		}
		remark := ""
		if r.Remark.Valid {
			remark = r.Remark.String
		}
		items = append(items, dto.GrowthReportGrowthItem{
			Time:              r.RecordTime,
			Height:            h,
			Weight:            w,
			HeadCircumference: hc,
			Remark:            remark,
		})
	}

	resp.Markdown = ""
	resp.Data = dto.GrowthReportData{
		Baby: dto.GrowthReportBaby{
			BabyID:   b.BabyID.String(),
			Name:     b.Name,
			Gender:   b.Gender,
			Birthday: b.Birthday,
			Avatar:   b.Avatar,
		},
		Range: dto.GrowthReportRange{
			From: from,
			To:   to,
			Days: days,
		},
		Growth: dto.GrowthReportGrowth{
			Items: items,
		},
		Analysis: dto.GrowthReportAnalysis{
			Height:            analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.Height }),
			Weight:            analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.Weight }),
			HeadCircumference: analyzeGrowthMetric(items, func(it dto.GrowthReportGrowthItem) *float64 { return it.HeadCircumference }),
		},
	}

	if !aiAvailable() {
		resp.Markdown = buildGrowthReportFallbackMarkdown(resp.Data, req.Language)
		return resp, nil
	}

	systemPrompt, userPrompt := buildGrowthReportPrompts(resp.Data, req.Language)
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}
	md, err := global.AIX.StreamChat(ctx, messages, func(chunk string) {})
	if err != nil {
		global.Log.Error(err)
		resp.Markdown = buildGrowthReportFallbackMarkdown(resp.Data, req.Language)
		return resp, nil
	}
	resp.Markdown = strings.TrimSpace(md)
	if resp.Markdown == "" {
		resp.Markdown = buildGrowthReportFallbackMarkdown(resp.Data, req.Language)
	}
	return resp, nil
}

func buildGrowthReportPrompts(data dto.GrowthReportData, language string) (string, string) {
	lang := strings.TrimSpace(strings.ToLower(language))
	if lang == "" {
		lang = "zh"
	}
	systemPrompt := "你是一个专业的儿科医生助手，擅长分析婴幼儿身高、体重、头围的变化趋势，并给出可执行的喂养与护理建议。请用中文输出。"
	if lang == "en" {
		systemPrompt = "You are a professional pediatric assistant. You analyze infant growth data (height, weight, head circumference), provide clear trend interpretation and actionable care/feeding suggestions. Output in English."
	}

	b, _ := json.Marshal(data)
	userPrompt := fmt.Sprintf(
		`请根据以下 JSON 数据生成一份“宝宝发育诊断与趋势预测报告”，输出 Markdown。

要求：
1) 不要臆测未提供的数据；若数据不足，请明确说明“数据不足以判断”，并给出补充记录建议。
2) 先给结论摘要（3-6 条要点），再给详细分析。
3) 分别对 身高/体重/头围 给出：最近变化、增长速度（周/月）、未来 30 天趋势预测（仅作参考）。
4) 给出 5-8 条可执行建议（喂养、作息、记录习惯），以及 3 道高营养辅食建议（用列表）。
5) 语言友好，避免制造焦虑。

JSON：
%s`, string(b),
	)
	if lang == "en" {
		userPrompt = fmt.Sprintf(
			`Generate a "Growth Assessment & 30-Day Trend Forecast Report" in Markdown based on the JSON below.

Requirements:
1) Do not invent missing data. If insufficient, explicitly say so and suggest what to record.
2) Provide a short summary (3-6 bullets), then details.
3) For height/weight/head circumference: recent change, growth rate (per week/month), and a 30-day forecast (informational only).
4) Provide 5-8 actionable suggestions and 3 nutrient-dense complementary food ideas.
5) Keep the tone supportive and non-alarming.

JSON:
%s`, string(b),
		)
	}
	return systemPrompt, userPrompt
}

func buildGrowthReportFallbackMarkdown(data dto.GrowthReportData, language string) string {
	lang := strings.TrimSpace(strings.ToLower(language))
	if lang == "" {
		lang = "zh"
	}
	if lang == "en" {
		return fmt.Sprintf(
			"# Growth Assessment Report\n\n## Baby\n- Name: %s\n- Gender: %s\n\n## Data Range\n- Days: %d\n\n## Summary\n- This report is generated without LLM due to a temporary AI error.\n\n## Trend (Informational)\n- Height: points=%d, delta=%s, per_week=%s, per_month=%s, predict_30=%s\n- Weight: points=%d, delta=%s, per_week=%s, per_month=%s, predict_30=%s\n- Head circumference: points=%d, delta=%s, per_week=%s, per_month=%s, predict_30=%s\n",
			data.Baby.Name,
			data.Baby.Gender,
			data.Range.Days,
			data.Analysis.Height.Points, fmtFloatPtr(data.Analysis.Height.Delta), fmtFloatPtr(data.Analysis.Height.PerWeek), fmtFloatPtr(data.Analysis.Height.PerMonth), fmtFloatPtr(data.Analysis.Height.Predict30),
			data.Analysis.Weight.Points, fmtFloatPtr(data.Analysis.Weight.Delta), fmtFloatPtr(data.Analysis.Weight.PerWeek), fmtFloatPtr(data.Analysis.Weight.PerMonth), fmtFloatPtr(data.Analysis.Weight.Predict30),
			data.Analysis.HeadCircumference.Points, fmtFloatPtr(data.Analysis.HeadCircumference.Delta), fmtFloatPtr(data.Analysis.HeadCircumference.PerWeek), fmtFloatPtr(data.Analysis.HeadCircumference.PerMonth), fmtFloatPtr(data.Analysis.HeadCircumference.Predict30),
		)
	}
	return fmt.Sprintf(
		"# 宝宝发育诊断与趋势预测报告\n\n## 宝宝信息\n- 名称：%s\n- 性别：%s\n\n## 数据范围\n- 天数：%d\n\n## 摘要\n- 当前报告为系统基础版（AI 暂不可用时生成）。\n\n## 趋势（仅供参考）\n- 身高：点数=%d，增量=%s，周增速=%s，月增速=%s，30天预测=%s\n- 体重：点数=%d，增量=%s，周增速=%s，月增速=%s，30天预测=%s\n- 头围：点数=%d，增量=%s，周增速=%s，月增速=%s，30天预测=%s\n",
		data.Baby.Name,
		data.Baby.Gender,
		data.Range.Days,
		data.Analysis.Height.Points, fmtFloatPtr(data.Analysis.Height.Delta), fmtFloatPtr(data.Analysis.Height.PerWeek), fmtFloatPtr(data.Analysis.Height.PerMonth), fmtFloatPtr(data.Analysis.Height.Predict30),
		data.Analysis.Weight.Points, fmtFloatPtr(data.Analysis.Weight.Delta), fmtFloatPtr(data.Analysis.Weight.PerWeek), fmtFloatPtr(data.Analysis.Weight.PerMonth), fmtFloatPtr(data.Analysis.Weight.Predict30),
		data.Analysis.HeadCircumference.Points, fmtFloatPtr(data.Analysis.HeadCircumference.Delta), fmtFloatPtr(data.Analysis.HeadCircumference.PerWeek), fmtFloatPtr(data.Analysis.HeadCircumference.PerMonth), fmtFloatPtr(data.Analysis.HeadCircumference.Predict30),
	)
}

func fmtFloatPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *v)
}

func analyzeGrowthMetric(items []dto.GrowthReportGrowthItem, pick func(it dto.GrowthReportGrowthItem) *float64) dto.GrowthMetricAnalysis {
	var firstTime int64
	var lastTime int64
	var firstVal float64
	var lastVal float64
	points := 0
	hasFirst := false
	hasLast := false

	for _, it := range items {
		v := pick(it)
		if v == nil {
			continue
		}
		points++
		if !hasFirst {
			firstTime = it.Time
			firstVal = *v
			hasFirst = true
		}
		lastTime = it.Time
		lastVal = *v
		hasLast = true
	}

	a := dto.GrowthMetricAnalysis{
		Points:    points,
		FirstTime: firstTime,
		LastTime:  lastTime,
	}
	if !hasFirst || !hasLast {
		return a
	}
	a.First = ptrFloat64(firstVal)
	a.Last = ptrFloat64(lastVal)

	if points < 2 {
		return a
	}
	deltaT := lastTime - firstTime
	if deltaT <= 0 {
		return a
	}
	days := float64(deltaT) / 86400000.0
	if days <= 0 {
		return a
	}
	delta := lastVal - firstVal
	perDay := delta / days
	a.Delta = ptrFloat64(delta)
	a.PerWeek = ptrFloat64(perDay * 7.0)
	a.PerMonth = ptrFloat64(perDay * 30.0)
	a.Predict30 = ptrFloat64(lastVal + perDay*30.0)
	return a
}

func ptrFloat64(v float64) *float64 {
	return &v
}
