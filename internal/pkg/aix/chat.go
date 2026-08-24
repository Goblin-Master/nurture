package aix

import (
	"context"
	"strings"

	"nurture/internal/config"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ChatMessage 对话消息
type ChatMessage struct {
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	Images    []string `json:"images,omitempty"`
	Timestamp int64    `json:"timestamp"`
}

func newChatModel(cfg config.ChatModel) (llms.Model, error) {
	return openai.New(
		openai.WithModel(cfg.Model),
		openai.WithToken(cfg.APIKey),
		openai.WithBaseURL(cfg.BaseURL),
	)
}

// StreamChat 流式对话
func (a *AIX) StreamChat(ctx context.Context, messages []llms.MessageContent,
	streamFunc func(chunk string)) (string, error) {
	if !a.ChatEnabled() {
		return "", ErrChatDisabled
	}

	var fullResponse strings.Builder

	_, err := a.chatModel.GenerateContent(ctx, messages,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			text := string(chunk)
			fullResponse.WriteString(text)
			streamFunc(text)
			return nil
		}),
	)

	if err != nil {
		return "", err
	}

	return fullResponse.String(), nil
}

// BuildMessages 构建消息（含历史 + RAG 上下文）
func (a *AIX) BuildMessages(history []ChatMessage, userMessage string,
	images []string, ragContext string, extraContext string) []llms.MessageContent {

	var messages []llms.MessageContent

	// 系统提示词
	systemPrompt := "你是一个稚慧云小助手。"
	if extraContext != "" {
		systemPrompt += "\n\n以下是宝宝相关系统数据：\n" + extraContext +
			"\n\n请结合以上数据回答用户问题。"
	}
	if ragContext != "" {
		systemPrompt += "\n\n以下是相关参考资料：\n" + ragContext +
			"\n\n请基于以上资料回答用户问题。"
	}
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt))

	// 历史消息
	for _, msg := range history {
		if msg.Role == "user" {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, msg.Content))
		} else {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, msg.Content))
		}
	}

	// 当前用户消息
	var parts []llms.ContentPart

	// 1. 添加图片 (如果有)
	for _, imgURL := range images {
		parts = append(parts, llms.ImageURLPart(imgURL))
	}

	// 2. 添加文本
	parts = append(parts, llms.TextPart(userMessage))

	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: parts,
	})

	return messages
}
