package test

import (
	"context"
	"errors"
	"strings"
	"testing"

	aidto "nurture/internal/ai/dto"
	ailogic "nurture/internal/ai/logic"
	airepo "nurture/internal/ai/repo"
	"nurture/internal/config"
	"nurture/internal/pkg/aix"

	"github.com/tmc/langchaingo/schema"
)

type aiRepoFake struct {
	fullHistory    []aix.ChatMessage
	fullHistoryErr error
}

func (f *aiRepoFake) AddDocument(context.Context, string, string) error {
	return nil
}

func (f *aiRepoFake) SimilaritySearch(context.Context, string, []string, int) ([]schema.Document, error) {
	return nil, nil
}

func (f *aiRepoFake) GetFullHistory(context.Context, string, string) ([]aix.ChatMessage, error) {
	if f.fullHistoryErr != nil {
		return nil, f.fullHistoryErr
	}
	return f.fullHistory, nil
}

func (f *aiRepoFake) GetRecentHistory(context.Context, string, string, int) ([]aix.ChatMessage, error) {
	return nil, nil
}

func (f *aiRepoFake) SaveMessage(context.Context, string, string, aix.ChatMessage) error {
	return nil
}

type growthReaderFake struct {
	baby    ailogic.BabyProfile
	babyErr error
	rows    []ailogic.GrowthRecord
	rowsErr error
}

func (f *growthReaderFake) GetBabyByIDAndUser(context.Context, string, string) (ailogic.BabyProfile, error) {
	if f.babyErr != nil {
		return ailogic.BabyProfile{}, f.babyErr
	}
	return f.baby, nil
}

func (f *growthReaderFake) ListGrowthRecordsByBabyIDBetween(context.Context, string, int64, int64) ([]ailogic.GrowthRecord, error) {
	if f.rowsErr != nil {
		return nil, f.rowsErr
	}
	return f.rows, nil
}

func TestAILogicChatStreamUnavailableEmitsError(t *testing.T) {
	l := ailogic.NewAILogic(&aiRepoFake{}, nil, config.AI{
		Chat: config.ChatModel{Enable: true},
	}, nil, false, nil)

	var events []aidto.SSEEvent
	err := l.ChatStream(t.Context(), "user-1", aidto.ChatStreamReq{
		SessionID: "session-1",
		Message:   "hello",
	}, func(event aidto.SSEEvent) {
		events = append(events, event)
	})

	if !errors.Is(err, ailogic.ErrChatStream) {
		t.Fatalf("ChatStream() error = %v, want %v", err, ailogic.ErrChatStream)
	}
	if len(events) != 1 || events[0].Type != "error" {
		t.Fatalf("events = %+v, want one error event", events)
	}
}

func TestAILogicUploadKnowledgeRejectsInvalidSpace(t *testing.T) {
	l := ailogic.NewAILogic(&aiRepoFake{}, nil, config.AI{}, nil, false, nil)

	err := l.UploadKnowledge(t.Context(), "user-1", aidto.KnowledgeUploadReq{
		SpaceType: "team",
		Content:   "content",
	})

	if !errors.Is(err, ailogic.ErrInvalidSpaceType) {
		t.Fatalf("UploadKnowledge() error = %v, want %v", err, ailogic.ErrInvalidSpaceType)
	}
}

func TestAILogicGetChatHistoryMapsRepoError(t *testing.T) {
	l := ailogic.NewAILogic(&aiRepoFake{fullHistoryErr: airepo.ErrHistoryGet}, nil, config.AI{}, nil, false, nil)

	_, err := l.GetChatHistory(t.Context(), "user-1", aidto.ChatHistoryReq{SessionID: "session-1"})

	if !errors.Is(err, ailogic.ErrDefault) {
		t.Fatalf("GetChatHistory() error = %v, want %v", err, ailogic.ErrDefault)
	}
}

func TestAILogicGrowthReportMapsBabyNotExist(t *testing.T) {
	l := ailogic.NewAILogic(&aiRepoFake{}, nil, config.AI{}, &growthReaderFake{
		babyErr: ailogic.ErrBabyNotExist,
	}, true, nil)

	_, err := l.GrowthReport(t.Context(), "user-1", aidto.GrowthReportReq{BabyID: "baby-1"})

	if !errors.Is(err, ailogic.ErrBabyNotExist) {
		t.Fatalf("GrowthReport() error = %v, want %v", err, ailogic.ErrBabyNotExist)
	}
}

func TestAILogicGrowthReportFallsBackWhenChatDisabled(t *testing.T) {
	height := 50.0
	weight := 3.5
	l := ailogic.NewAILogic(&aiRepoFake{}, nil, config.AI{}, &growthReaderFake{
		baby: ailogic.BabyProfile{
			BabyID:   "baby-1",
			Name:     "Mia",
			Gender:   "female",
			Birthday: 1704067200000,
			Avatar:   "avatar.png",
		},
		rows: []ailogic.GrowthRecord{
			{RecordTime: 1704067200000, Height: &height, Weight: &weight},
		},
	}, true, nil)

	resp, err := l.GrowthReport(t.Context(), "user-1", aidto.GrowthReportReq{BabyID: "baby-1"})

	if err != nil {
		t.Fatalf("GrowthReport() error = %v", err)
	}
	if resp.Data.Baby.BabyID != "baby-1" || len(resp.Data.Growth.Items) != 1 {
		t.Fatalf("GrowthReport() data = %+v, want baby and one growth item", resp.Data)
	}
	if !strings.Contains(resp.Markdown, "Mia") {
		t.Fatalf("fallback markdown = %q, want baby name", resp.Markdown)
	}
}
