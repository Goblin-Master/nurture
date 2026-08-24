package aix

import (
	"errors"
	"nurture/internal/config"
	"testing"
)

func TestNewAIXInitializesEnabledCapabilitiesOnly(t *testing.T) {
	ai, err := NewAIX(config.AI{
		Chat: config.ChatModel{Enable: false},
		Embedding: config.EmbeddingModel{
			Enable:  true,
			Model:   "BAAI/bge-m3",
			BaseURL: "https://api.siliconflow.cn/v1",
		},
	}, nil, "")
	if err != nil {
		t.Fatalf("NewAIX() error = %v", err)
	}
	if ai.ChatEnabled() {
		t.Fatalf("ChatEnabled() = true, want false")
	}
	if !ai.EmbeddingEnabled() {
		t.Fatalf("EmbeddingEnabled() = false, want true")
	}
	if _, err := ai.StreamChat(nil, nil, nil); !errors.Is(err, ErrChatDisabled) {
		t.Fatalf("StreamChat() error = %v, want ErrChatDisabled", err)
	}
}

func TestNewAIXWithAllCapabilitiesDisabled(t *testing.T) {
	ai, err := NewAIX(config.AI{}, nil, "")
	if err != nil {
		t.Fatalf("NewAIX() error = %v", err)
	}
	if ai.ChatEnabled() {
		t.Fatalf("ChatEnabled() = true, want false")
	}
	if ai.EmbeddingEnabled() {
		t.Fatalf("EmbeddingEnabled() = true, want false")
	}
}
