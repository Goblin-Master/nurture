package repo

import (
	"context"
	"encoding/json"
	"fmt"
	aiconstant "nurture/internal/ai/constant"
	"nurture/internal/pkg/aix"
	"nurture/internal/pkg/zapx"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tmc/langchaingo/schema"
	"go.uber.org/zap"
)

type IAIRepo interface {
	AddDocument(ctx context.Context, collectionName, content string) error
	SimilaritySearch(ctx context.Context, query string, collections []string, topK int) ([]schema.Document, error)
	GetFullHistory(ctx context.Context, userID, sessionID string) ([]aix.ChatMessage, error)
	GetRecentHistory(ctx context.Context, userID, sessionID string, limit int) ([]aix.ChatMessage, error)
	SaveMessage(ctx context.Context, userID, sessionID string, msg aix.ChatMessage) error
}

type AIRepo struct {
	ai  *aix.AIX
	rdb redis.Cmdable
	log *zap.SugaredLogger
}

func NewAIRepo(ai *aix.AIX, rdb redis.Cmdable, log *zap.SugaredLogger) *AIRepo {
	return &AIRepo{
		ai:  ai,
		rdb: rdb,
		log: zapx.OrNop(log),
	}
}

var _ IAIRepo = (*AIRepo)(nil)

func (r *AIRepo) logError(err error) {
	if err != nil {
		r.log.Error(err)
	}
}

func (r *AIRepo) AddDocument(ctx context.Context, collectionName, content string) error {
	if r.ai == nil || !r.ai.EmbeddingEnabled() {
		return ErrDocumentAdd
	}
	err := r.ai.AddDocument(ctx, collectionName, content)
	if err != nil {
		r.logError(err)
		return ErrDocumentAdd
	}
	return nil
}

func (r *AIRepo) SimilaritySearch(ctx context.Context, query string,
	collections []string, topK int) ([]schema.Document, error) {
	if r.ai == nil || !r.ai.EmbeddingEnabled() {
		return nil, ErrDocumentSearch
	}
	docs, err := r.ai.SimilaritySearch(ctx, query, collections, topK)
	if err != nil {
		r.logError(err)
		return nil, ErrDocumentSearch
	}
	return docs, nil
}

// GetFullHistory 获取完整对话历史
func (r *AIRepo) GetFullHistory(ctx context.Context, userID, sessionID string) ([]aix.ChatMessage, error) {
	if r.rdb == nil {
		// 未初始化 Redis,直接返回空，方便测试
		return nil, nil
	}
	key := fmt.Sprintf(aiconstant.ChatHistoryKey, userID, sessionID)
	// 获取所有消息
	result, err := r.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		r.logError(err)
		return nil, ErrHistoryGet
	}

	var messages []aix.ChatMessage
	for _, item := range result {
		var msg aix.ChatMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			r.log.Errorf("Unmarshal message failed: %v", err)
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetRecentHistory 获取最近 N 条历史记录
func (r *AIRepo) GetRecentHistory(ctx context.Context, userID, sessionID string, limit int) ([]aix.ChatMessage, error) {
	if r.rdb == nil {
		return nil, nil
	}
	key := fmt.Sprintf(aiconstant.ChatHistoryKey, userID, sessionID)
	// 获取最后 limit 条
	// LRange start stop: 0 is first, -1 is last.
	// To get last N: start = -N, stop = -1
	start := int64(-limit)
	result, err := r.rdb.LRange(ctx, key, start, -1).Result()
	if err != nil {
		r.logError(err)
		return nil, ErrHistoryGet
	}

	var messages []aix.ChatMessage
	for _, item := range result {
		var msg aix.ChatMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			r.log.Errorf("Unmarshal message failed: %v", err)
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// SaveMessage 保存消息到 Redis List
func (r *AIRepo) SaveMessage(ctx context.Context, userID, sessionID string, msg aix.ChatMessage) error {
	if r.rdb == nil {
		return nil
	}
	key := fmt.Sprintf(aiconstant.ChatHistoryKey, userID, sessionID)
	data, err := json.Marshal(msg)
	if err != nil {
		r.logError(err)
		return ErrHistorySave
	}

	// 存入 List 尾部
	if err := r.rdb.RPush(ctx, key, data).Err(); err != nil {
		r.logError(err)
		return ErrHistorySave
	}

	// 刷新过期时间
	if err := r.rdb.Expire(ctx, key, time.Duration(aiconstant.HistoryTTL)*time.Second).Err(); err != nil {
		r.logError(err)
		return ErrHistorySave
	}
	return nil
}
