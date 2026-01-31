package test

import (
	"context"
	"fmt"
	"nurture/internal/config"
	"nurture/internal/pkg/aix"
	"nurture/internal/pkg/jwtx"
	"testing"
	"time"
)

func init() {
	// 加载配置
	config.LoadConfig()
}

func TestStreamChat(t *testing.T) {
	// 1. 生成测试 Token (虽然这个测试是直接调用 pkg 方法不涉及 Handler，但为了符合规范演示 Token 生成)
	token, err := jwtx.GenTestToken("test_user_id", jwtx.COMMON_USER)
	if err != nil {
		t.Fatalf("GenTestToken failed: %v", err)
	}
	t.Logf("Generated Test Token: %s", token)

	// 2. 初始化 AIX (使用 config 加载的配置)
	cfg := config.Conf.AI
	ai, err := aix.NewAIX(cfg, nil, "")
	if err != nil {
		t.Fatalf("NewAIX failed: %v", err)
	}

	// 3. 构建测试消息
	userMessage := "你好，请用一句话介绍一下你自己，并告诉我今天是星期几（假设今天是2026年1月31日）"
	messages := ai.BuildMessages(nil, userMessage, nil, "")

	fmt.Printf("\n=== Start Streaming Chat ===\n")
	fmt.Printf("User: %s\n", userMessage)
	fmt.Printf("AI: ")

	// 4. 执行流式对话并实时打印
	fullResponse, err := ai.StreamChat(context.Background(), messages, func(chunk string) {
		// 实时打印每个 chunk，不换行，模拟打字机效果
		fmt.Print(chunk)
		// 稍微延时以便观察流式效果 (仅测试用)
		time.Sleep(50 * time.Millisecond)
	})
	fmt.Printf("\n=== End Streaming Chat ===\n\n")

	if err != nil {
		t.Errorf("StreamChat failed: %v", err)
	}

	if fullResponse == "" {
		t.Error("Response is empty")
	}
	t.Logf("Full Response length: %d", len(fullResponse))
}
