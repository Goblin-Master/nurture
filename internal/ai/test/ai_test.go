package test

import (
	"context"
	"fmt"
	aiconstant "nurture/internal/ai/constant"
	aidto "nurture/internal/ai/dto"
	ailogic "nurture/internal/ai/logic"
	airepo "nurture/internal/ai/repo"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/pkg/jwtx"
	"sync"
	"testing"
	"time"
)

var (
	globalInitOnce  sync.Once
	globalInitPanic any
)

func loadConfig(t *testing.T) {
	t.Helper()
	config.LoadConfig()
}

func initGlobal(t *testing.T) {
	t.Helper()
	globalInitOnce.Do(func() {
		defer func() {
			globalInitPanic = recover()
		}()
		global.Init()
	})
	if globalInitPanic != nil {
		t.Fatalf("global.Init failed: %v", globalInitPanic)
	}
}

func skipAIChatIntegration(t *testing.T) {
	t.Helper()
	if !config.Conf.AI.Chat.Enable {
		t.Skip("skip ai chat integration test: ai.chat.enable=false")
	}
}

func skipAIEmbeddingIntegration(t *testing.T) {
	t.Helper()
	if !config.Conf.AI.Embedding.Enable {
		t.Skip("skip ai embedding integration test: ai.embedding.enable=false")
	}
}

func skipAIVectorIntegration(t *testing.T) {
	t.Helper()
	skipAIEmbeddingIntegration(t)
	if !config.Conf.DB.Enable {
		t.Skip("skip ai vector integration test: db.enable=false")
	}
}

func TestStreamChat(t *testing.T) {
	loadConfig(t)
	skipAIChatIntegration(t)
	initGlobal(t)

	// 1. 生成测试 Token (虽然这个测试是直接调用 pkg 方法不涉及 Handler，但为了符合规范演示 Token 生成)
	token, err := jwtx.GenTestToken("test_user_id", jwtx.COMMON_USER)
	if err != nil {
		t.Fatalf("GenTestToken failed: %v", err)
	}
	t.Logf("Generated Test Token: %s", token)

	// 2. 初始化 Logic
	aiRepo := airepo.NewAIRepo(global.AIX, global.RDB, global.Log)
	aiLogic := ailogic.NewAILogic(aiRepo, global.AIX, config.Conf.AI, nil, config.Conf.DB.Enable && global.DB != nil, global.Log)

	// 3. 准备请求数据 (SessionID 为空，测试自动生成)
	req := aidto.ChatStreamReq{
		SessionID: "",
		Message:   "你好，请用一句话介绍一下你自己，并告诉我今天是星期几",
	}

	fmt.Printf("\n=== Start Streaming Chat (Logic Layer) ===\n")
	fmt.Printf("User: %s\n", req.Message)
	fmt.Printf("AI: ")

	// 4. 调用 Logic 层 ChatStream
	userID := "test_user_id_from_token"

	err = aiLogic.ChatStream(context.Background(), userID, req, func(event aidto.SSEEvent) {
		// 实时打印每个 chunk
		if event.Type == "content" {
			fmt.Print(event.Content)
		}
		time.Sleep(50 * time.Millisecond)
	})
	fmt.Printf("\n=== End Streaming Chat ===\n")

	if err != nil {
		t.Errorf("ChatStream failed: %v", err)
	}
}

func TestEmbedding(t *testing.T) {
	loadConfig(t)
	skipAIEmbeddingIntegration(t)
	initGlobal(t)

	// 1. 测试文本
	text := "小王是个大混蛋"
	fmt.Printf("\n=== Testing Embedding (Model: %s) ===\n", config.Conf.AI.Embedding.Model)
	fmt.Printf("Input text: %s\n", text)

	// 2. 调用向量化
	vector, err := global.AIX.EmbedDocument(context.Background(), text)
	if err != nil {
		t.Fatalf("EmbedDocument failed: %v", err)
	}

	// 3. 验证结果
	fmt.Printf("Vector dimension: %d\n", len(vector))
	if len(vector) == 0 {
		t.Error("Vector is empty")
	}
	preview := min(5, len(vector))
	t.Logf("Vector (first %d elements): %v", preview, vector[:preview])
}

func TestSimilaritySearch(t *testing.T) {
	loadConfig(t)
	skipAIVectorIntegration(t)
	initGlobal(t)

	aiRepo := airepo.NewAIRepo(global.AIX, global.RDB, global.Log)
	userID := "test_user_id"
	collectionName := fmt.Sprintf(aiconstant.CollectionUserPrefix, userID)

	// 1. 添加文档
	content := "小王是个大混蛋"
	ctx := context.Background()
	fmt.Printf("\n=== Testing Similarity Search ===\n")
	fmt.Printf("Adding document: %s\n", content)
	err := aiRepo.AddDocument(ctx, collectionName, content)
	if err != nil {
		t.Fatalf("AddDocument failed: %v", err)
	}

	// 2. 检索
	query := "小王是什么"
	fmt.Printf("Querying: %s\n", query)
	docs, err := aiRepo.SimilaritySearch(ctx, query, []string{collectionName}, config.Conf.AI.Retrieval.DefaultTopK)
	if err != nil {
		t.Fatalf("SimilaritySearch failed: %v", err)
	}

	// 3. 验证
	if len(docs) == 0 {
		t.Fatal("No documents found")
	}
	fmt.Printf("Found %d documents\n", len(docs))
	fmt.Printf("Top match: %s\n", docs[0].PageContent)
}
