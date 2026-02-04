package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/global"
	"nurture/internal/pkg/aix"
	"time"

	"github.com/tmc/langchaingo/schema"
)

type IAIRepo interface {
	AddDocument(ctx context.Context, collectionName, content string) error
	SimilaritySearch(ctx context.Context, query string, collections []string, topK int) ([]schema.Document, error)
	GetFullHistory(ctx context.Context, userID, sessionID string) ([]aix.ChatMessage, error)
	GetRecentHistory(ctx context.Context, userID, sessionID string, limit int) ([]aix.ChatMessage, error)
	SaveMessage(ctx context.Context, userID, sessionID string, msg aix.ChatMessage) error
}

type AIRepo struct{}

func NewAIRepo() *AIRepo {
	return &AIRepo{}
}

var _ IAIRepo = (*AIRepo)(nil)

func (r *AIRepo) AddDocument(ctx context.Context, collectionName, content string) error {
	err := global.AIX.AddDocument(ctx, collectionName, content)
	if err != nil {
		global.Log.Error(err)
		return ErrDocumentAdd
	}
	return nil
}

func (r *AIRepo) SimilaritySearch(ctx context.Context, query string,
	collections []string, topK int) ([]schema.Document, error) {
	docs, err := global.AIX.SimilaritySearch(ctx, query, collections, config.Conf.AI.Retrieval.DefaultTopK)
	if err != nil {
		global.Log.Error(err)
		return nil, ErrDocumentSearch
	}
	return docs, nil
}

// GetFullHistory 获取完整对话历史
func (r *AIRepo) GetFullHistory(ctx context.Context, userID, sessionID string) ([]aix.ChatMessage, error) {
	if global.RDB == nil {
		// 未初始化 Redis,直接返回空，方便测试
		return nil, nil
	}
	key := fmt.Sprintf(constant.CHAT_HISTORY_KEY, userID, sessionID)
	// 获取所有消息
	result, err := global.RDB.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []aix.ChatMessage
	for _, item := range result {
		var msg aix.ChatMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			global.Log.Errorf("Unmarshal message failed: %v", err)
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetRecentHistory 获取最近 N 条历史记录
func (r *AIRepo) GetRecentHistory(ctx context.Context, userID, sessionID string, limit int) ([]aix.ChatMessage, error) {
	if global.RDB == nil {
		return nil, nil
	}
	key := fmt.Sprintf(constant.CHAT_HISTORY_KEY, userID, sessionID)
	// 获取最后 limit 条
	// LRange start stop: 0 is first, -1 is last.
	// To get last N: start = -N, stop = -1
	start := int64(-limit)
	result, err := global.RDB.LRange(ctx, key, start, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []aix.ChatMessage
	for _, item := range result {
		var msg aix.ChatMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			global.Log.Errorf("Unmarshal message failed: %v", err)
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// SaveMessage 保存消息到 Redis List
func (r *AIRepo) SaveMessage(ctx context.Context, userID, sessionID string, msg aix.ChatMessage) error {
	if global.RDB == nil {
		return nil
	}
	key := fmt.Sprintf(constant.CHAT_HISTORY_KEY, userID, sessionID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 存入 List 尾部
	if err := global.RDB.RPush(ctx, key, data).Err(); err != nil {
		return err
	}

	// 刷新过期时间
	return global.RDB.Expire(ctx, key, time.Duration(constant.HISTORY_TTL)*time.Second).Err()
}
