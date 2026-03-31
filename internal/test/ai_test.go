package test

import (
	"context"
	"fmt"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/pkg/aix"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/repo"
	"os"
	"testing"
	"time"
)

func init() {
	// 加载配置
	config.LoadConfig()
}

func TestStreamChat(t *testing.T) {
	if os.Getenv("NURTURE_RUN_AI_TESTS") != "1" {
		t.Skip("skip ai integration test: set NURTURE_RUN_AI_TESTS=1 to run")
	}
	// 1. 生成测试 Token (虽然这个测试是直接调用 pkg 方法不涉及 Handler，但为了符合规范演示 Token 生成)
	token, err := jwtx.GenTestToken("test_user_id", jwtx.COMMON_USER)
	if err != nil {
		t.Fatalf("GenTestToken failed: %v", err)
	}
	t.Logf("Generated Test Token: %s", token)

	// 2. 初始化 Global AIX (Logic 层依赖 global.AIX)
	// 注意：这里 RDB 传 nil，因为 repo 层的 dummy 实现还没真正用 Redis，如果有依赖需要 Mock
	global.AIX, err = aix.NewAIX(config.Conf.AI, nil, "")
	if err != nil {
		t.Fatalf("NewAIX failed: %v", err)
	}

	// 3. 初始化 Logic
	commonLogic := logic.NewCommonLogic()

	// 4. 准备请求数据 (SessionID 为空，测试自动生成)
	req := dto.ChatStreamReq{
		SessionID: "",
		Message:   "你好，请用一句话介绍一下你自己，并告诉我今天是星期几",
	}

	fmt.Printf("\n=== Start Streaming Chat (Logic Layer) ===\n")
	fmt.Printf("User: %s\n", req.Message)
	fmt.Printf("AI: ")

	// 5. 调用 Logic 层 ChatStream
	// 模拟从 Token 解析出的 UserID
	userID := "test_user_id_from_token"

	err = commonLogic.ChatStream(context.Background(), userID, req, func(event dto.SSEEvent) {
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
	if os.Getenv("NURTURE_RUN_AI_TESTS") != "1" {
		t.Skip("skip ai integration test: set NURTURE_RUN_AI_TESTS=1 to run")
	}
	// 1. 初始化 AIX
	ai, err := aix.NewAIX(config.Conf.AI, nil, "")
	if err != nil {
		t.Fatalf("NewAIX failed: %v", err)
	}

	// 2. 测试文本
	text := "小王是个大混蛋"
	fmt.Printf("\n=== Testing Embedding (Model: %s) ===\n", config.Conf.AI.Embedding.Model)
	fmt.Printf("Input text: %s\n", text)

	// 3. 调用向量化
	vector, err := ai.EmbedDocument(context.Background(), text)
	if err != nil {
		t.Fatalf("EmbedDocument failed: %v", err)
	}

	// 4. 验证结果
	fmt.Printf("Vector dimension: %d\n", len(vector))
	if len(vector) == 0 {
		t.Error("Vector is empty")
	}
	// 智谱 embedding-3 默认维度通常是 2048，embedding-2 是 1024
	t.Logf("Vector (first 5 elements): %v", vector[:5])
}

func TestSimilaritySearch(t *testing.T) {
	if os.Getenv("NURTURE_RUN_AI_TESTS") != "1" {
		t.Skip("skip ai integration test: set NURTURE_RUN_AI_TESTS=1 to run")
	}
	// 1. 初始化 (确保 Global AIX 被正确初始化，且包含 DB 连接)
	var err error
	// 必须传入有效的 DSN，否则无法连接向量库
	global.AIX, err = aix.NewAIX(config.Conf.AI, nil, config.Conf.DB.DSN())
	if err != nil {
		t.Fatalf("NewAIX failed: %v", err)
	}

	aiRepo := repo.NewAIRepo()
	userID := "test_user_id"
	collectionName := fmt.Sprintf(constant.COLLECTION_USER_PREFIX, userID)

	// 2. 添加文档
	content := "小王是个大混蛋"
	ctx := context.Background()
	fmt.Printf("\n=== Testing Similarity Search ===\n")
	fmt.Printf("Adding document: %s\n", content)
	err = aiRepo.AddDocument(ctx, collectionName, content)
	if err != nil {
		t.Fatalf("AddDocument failed: %v", err)
	}

	// 3. 检索
	query := "小王是什么"
	fmt.Printf("Querying: %s\n", query)
	docs, err := aiRepo.SimilaritySearch(ctx, query, []string{collectionName}, config.Conf.AI.Retrieval.DefaultTopK)
	if err != nil {
		t.Fatalf("SimilaritySearch failed: %v", err)
	}

	// 4. 验证
	if len(docs) == 0 {
		t.Fatal("No documents found")
	}
	fmt.Printf("Found %d documents\n", len(docs))
	fmt.Printf("Top match: %s\n", docs[0].PageContent)
}
