package repo

import (
	"context"
	"nurture/internal/pkg/aix"

	"github.com/tmc/langchaingo/schema"
)

type IAIRepo interface {
	AddDocuments(ctx context.Context, collectionName, content string) error
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

func (r *AIRepo) AddDocuments(ctx context.Context, collectionName, content string) error {
	// Dummy implementation for now
	return nil
}

func (r *AIRepo) SimilaritySearch(ctx context.Context, query string,
	collections []string, topK int) ([]schema.Document, error) {
	// Dummy implementation for now
	return nil, nil
}

func (r *AIRepo) GetFullHistory(ctx context.Context, userID, sessionID string) ([]aix.ChatMessage, error) {
	// Dummy implementation for now
	return []aix.ChatMessage{}, nil
}

func (r *AIRepo) GetRecentHistory(ctx context.Context, userID, sessionID string, limit int) ([]aix.ChatMessage, error) {
	// Dummy implementation for now
	return []aix.ChatMessage{}, nil
}

func (r *AIRepo) SaveMessage(ctx context.Context, userID, sessionID string, msg aix.ChatMessage) error {
	// Dummy implementation for now
	return nil
}
