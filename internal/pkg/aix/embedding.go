package aix

import (
	"context"
	"fmt"
	"nurture/internal/config"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	ProviderOpenAI            = "openai"
	ProviderSiliconFlow       = "siliconflow"
	ProviderZhipu             = "zhipu"
	defaultSiliconFlowBaseURL = "https://api.siliconflow.cn/v1"
	defaultZhipuBaseURL       = "https://open.bigmodel.cn/api/paas/v4"
)

// newEmbeddingModel 初始化 Embedding 模型
func newEmbeddingModel(cfg config.EmbeddingModel, providers map[string]EmbeddingProviderFactory) (embeddings.Embedder, error) {
	provider := normalizeProvider(cfg.Provider)
	if provider == "" {
		provider = ProviderSiliconFlow
	}

	factory, ok := providers[provider]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedEmbeddingProvider, provider)
	}
	return factory(cfg)
}

func newSiliconFlowEmbedder(cfg config.EmbeddingModel) (embeddings.Embedder, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultSiliconFlowBaseURL
	}
	return newOpenAICompatibleEmbedder(cfg)
}

func newZhipuEmbedder(cfg config.EmbeddingModel) (embeddings.Embedder, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultZhipuBaseURL
	}
	return newOpenAICompatibleEmbedder(cfg)
}

func newOpenAICompatibleEmbedder(cfg config.EmbeddingModel) (embeddings.Embedder, error) {
	llm, err := openai.New(
		openai.WithToken(cfg.APIKey),
		openai.WithBaseURL(cfg.BaseURL),
		openai.WithEmbeddingModel(cfg.Model),
	)
	if err != nil {
		return nil, err
	}
	return embeddings.NewEmbedder(llm)
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// EmbedQuery 文本向量化 (单条)
func (a *AIX) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if !a.EmbeddingEnabled() {
		return nil, ErrEmbeddingDisabled
	}
	return a.embedder.EmbedQuery(ctx, text)
}

// EmbedDocument 文本向量化 (单条)
func (a *AIX) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	if !a.EmbeddingEnabled() {
		return nil, ErrEmbeddingDisabled
	}
	embeddings, err := a.embedder.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, nil
	}
	return embeddings[0], nil
}
