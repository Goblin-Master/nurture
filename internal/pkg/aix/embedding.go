package aix

import (
	"context"
	"fmt"
	"nurture/internal/config"

	"github.com/tmc/langchaingo/embeddings"
)

// newEmbeddingModel 初始化 Embedding 模型
func newEmbeddingModel(cfg config.EmbeddingModel) (embeddings.Embedder, error) {
	fmt.Printf("DEBUG: Init Embedding Model: %s, BaseURL: %s\n", cfg.Model, cfg.BaseURL)

	// 使用自定义的智谱 Embedder，绕过 langchaingo 的兼容性问题
	return NewZhipuEmbedder(cfg), nil
}

// EmbedQuery 文本向量化 (单条)
func (a *AIX) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return a.embedder.EmbedQuery(ctx, text)
}

// EmbedDocument 文本向量化 (单条)
func (a *AIX) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := a.embedder.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, nil
	}
	return embeddings[0], nil
}
