package aix

import (
	"nurture/internal/config"

	"github.com/go-redis/redis/v8"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

// AIX 封装所有 AI 相关功能
type AIX struct {
	chatModel llms.Model
	embedder  embeddings.Embedder
	rdb       redis.Cmdable
	pgConnURL string
	config    config.AI
}

// NewAIX 创建 AIX 实例
func NewAIX(cfg config.AI, rdb redis.Cmdable, pgConnURL string) (*AIX, error) {
	// 初始化 Chat 模型
	chatModel, err := newChatModel(cfg.Chat)
	if err != nil {
		return nil, err
	}

	// 初始化 Embedding 模型
	embedder, err := newEmbeddingModel(cfg.Embedding)
	if err != nil {
		return nil, err
	}

	return &AIX{
		chatModel: chatModel,
		embedder:  embedder,
		rdb:       rdb,
		pgConnURL: pgConnURL,
		config:    cfg,
	}, nil
}
