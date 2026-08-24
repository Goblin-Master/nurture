package aix

import (
	"context"
	"errors"
	"testing"

	"nurture/internal/config"

	"github.com/tmc/langchaingo/embeddings"
)

type stubEmbedder struct{}

func (stubEmbedder) EmbedDocuments(context.Context, []string) ([][]float32, error) {
	return nil, nil
}

func (stubEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return nil, nil
}

func TestNewEmbeddingModelProviderSelection(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{name: "default provider", want: ProviderSiliconFlow},
		{name: "explicit siliconflow", provider: ProviderSiliconFlow, want: ProviderSiliconFlow},
		{name: "explicit zhipu", provider: ProviderZhipu, want: ProviderZhipu},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := ""
			_, err := newEmbeddingModel(config.EmbeddingModel{Provider: tt.provider}, map[string]EmbeddingProviderFactory{
				ProviderSiliconFlow: func(config.EmbeddingModel) (embeddings.Embedder, error) {
					called = ProviderSiliconFlow
					return stubEmbedder{}, nil
				},
				ProviderZhipu: func(config.EmbeddingModel) (embeddings.Embedder, error) {
					called = ProviderZhipu
					return stubEmbedder{}, nil
				},
			})
			if err != nil {
				t.Fatalf("newEmbeddingModel() error = %v", err)
			}
			if called != tt.want {
				t.Fatalf("provider = %q, want %q", called, tt.want)
			}
		})
	}
}

func TestNewEmbeddingModelUnsupportedProvider(t *testing.T) {
	_, err := newEmbeddingModel(config.EmbeddingModel{Provider: "unknown"}, map[string]EmbeddingProviderFactory{})
	if !errors.Is(err, ErrUnsupportedEmbeddingProvider) {
		t.Fatalf("newEmbeddingModel() error = %v, want ErrUnsupportedEmbeddingProvider", err)
	}
}

func TestNewAIXUsesEmbeddingProviderOption(t *testing.T) {
	ai, err := NewAIX(config.AI{
		Embedding: config.EmbeddingModel{
			Enable:   true,
			Provider: "custom",
		},
	}, nil, "", WithEmbeddingProvider("custom", func(config.EmbeddingModel) (embeddings.Embedder, error) {
		return stubEmbedder{}, nil
	}))
	if err != nil {
		t.Fatalf("NewAIX() error = %v", err)
	}
	if ai.ChatEnabled() {
		t.Fatalf("ChatEnabled() = true, want false")
	}
	if !ai.EmbeddingEnabled() {
		t.Fatalf("EmbeddingEnabled() = false, want true")
	}
}
