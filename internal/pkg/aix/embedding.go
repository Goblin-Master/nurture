package aix

import (
	"context"
	"nurture/internal/config"

	"github.com/tmc/langchaingo/embeddings"
)

// newEmbeddingModel 初始化 Embedding 模型
func newEmbeddingModel(cfg config.EmbeddingModel) (embeddings.Embedder, error) {
	return NewSiliconFlowEmbedder(cfg), nil
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
