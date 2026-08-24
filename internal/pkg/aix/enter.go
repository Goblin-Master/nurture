package aix

import (
	"nurture/internal/config"

	"github.com/go-redis/redis/v8"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

type EmbeddingProviderFactory func(config.EmbeddingModel) (embeddings.Embedder, error)

type Option func(*options)

type options struct {
	embeddingProviders map[string]EmbeddingProviderFactory
}

// AIX 封装所有 AI 相关功能
type AIX struct {
	chatModel llms.Model
	embedder  embeddings.Embedder
	rdb       redis.Cmdable
	pgConnURL string
	config    config.AI
}

// NewAIX 创建 AIX 实例
func NewAIX(cfg config.AI, rdb redis.Cmdable, pgConnURL string, opts ...Option) (*AIX, error) {
	option := defaultOptions()
	for _, opt := range opts {
		opt(&option)
	}

	var chatModel llms.Model
	if cfg.Chat.Enable {
		var err error
		chatModel, err = newChatModel(cfg.Chat)
		if err != nil {
			return nil, err
		}
	}

	var embedder embeddings.Embedder
	if cfg.Embedding.Enable {
		var err error
		embedder, err = newEmbeddingModel(cfg.Embedding, option.embeddingProviders)
		if err != nil {
			return nil, err
		}
	}

	return &AIX{
		chatModel: chatModel,
		embedder:  embedder,
		rdb:       rdb,
		pgConnURL: pgConnURL,
		config:    cfg,
	}, nil
}

func (a *AIX) ChatEnabled() bool {
	return a != nil && a.chatModel != nil
}

func (a *AIX) EmbeddingEnabled() bool {
	return a != nil && a.embedder != nil
}

func WithEmbeddingProvider(provider string, factory EmbeddingProviderFactory) Option {
	return func(opts *options) {
		provider = normalizeProvider(provider)
		if provider == "" || factory == nil {
			return
		}
		opts.embeddingProviders[provider] = factory
	}
}

func defaultOptions() options {
	return options{
		embeddingProviders: map[string]EmbeddingProviderFactory{
			ProviderOpenAI:      newOpenAICompatibleEmbedder,
			ProviderSiliconFlow: newSiliconFlowEmbedder,
			ProviderZhipu:       newZhipuEmbedder,
		},
	}
}
