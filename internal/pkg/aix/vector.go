package aix

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// AddDocument 添加文档到知识库
func (a *AIX) AddDocument(ctx context.Context, collectionName string, content string) error {
	// 1. 文本分块
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(a.config.Chunking.ChunkSize),
		textsplitter.WithChunkOverlap(a.config.Chunking.ChunkOverlap),
	)
	chunks, err := splitter.SplitText(content)
	if err != nil {
		return err
	}

	// 2. 转换为 Document
	docs := make([]schema.Document, len(chunks))
	for i, chunk := range chunks {
		docs[i] = schema.Document{
			PageContent: chunk,
			Metadata:    map[string]any{},
		}
	}

	// 3. 创建向量存储
	store, err := a.newVectorStore(ctx, collectionName)
	if err != nil {
		return err
	}

	// 4. 添加文档
	_, err = store.AddDocuments(ctx, docs)
	return err
}

// SimilaritySearch 相似度搜索
func (a *AIX) SimilaritySearch(ctx context.Context, query string,
	collections []string, topK int) ([]schema.Document, error) {

	var allDocs []schema.Document

	for _, collName := range collections {
		store, err := a.newVectorStore(ctx, collName)
		if err != nil {
			continue
		}

		// 使用 ScoreThreshold 过滤
		docs, err := store.SimilaritySearch(ctx, query, topK,
			vectorstores.WithScoreThreshold(a.config.Retrieval.SimilarityThreshold),
		)
		if err != nil {
			continue
		}
		allDocs = append(allDocs, docs...)
	}

	return allDocs, nil
}

func (a *AIX) newVectorStore(ctx context.Context, collectionName string) (pgvector.Store, error) {
	return pgvector.New(
		ctx,
		pgvector.WithConnectionURL(a.pgConnURL),
		pgvector.WithEmbedder(a.embedder),
		pgvector.WithCollectionName(collectionName),
		pgvector.WithPreDeleteCollection(false),
	)
}
