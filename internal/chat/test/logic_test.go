package test

import (
	"context"
	"errors"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/dto"
	"nurture/internal/chat/event"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/repo"
	"testing"
	"time"

	"github.com/google/uuid"
)

type allowAllLimiter struct{}

func (allowAllLimiter) Allow(_ context.Context, _ string, limit int64, _ time.Duration) (bool, int64, error) {
	return true, limit, nil
}

type publisherFake struct {
	direct event.DirectMessage
	group  event.GroupMessage
	err    error
}

func (f *publisherFake) PublishDirect(_ context.Context, msg event.DirectMessage) error {
	f.direct = msg
	return f.err
}

func (f *publisherFake) PublishGroup(_ context.Context, msg event.GroupMessage) error {
	f.group = msg
	return f.err
}

type chatRepoFake struct {
	joinGroupErr     error
	getMemberRoleErr error

	saveDirectMessageErr error
	savedDirectMessage   savedDirectMessage
}

type savedDirectMessage struct {
	messageID  string
	fromUserID string
	toUserID   string
	msgType    string
	content    string
	now        int64
}

func (f *chatRepoFake) CreateGroup(context.Context, string, string, string, string, string, int32, int64) error {
	return nil
}

func (f *chatRepoFake) JoinGroup(context.Context, string, string, int64) error {
	return f.joinGroupErr
}

func (f *chatRepoFake) LeaveGroup(context.Context, string, string, int64) error {
	return nil
}

func (f *chatRepoFake) TransferOwner(context.Context, string, string, string, int64) error {
	return nil
}

func (f *chatRepoFake) DissolveGroup(context.Context, string, string, int64) error {
	return nil
}

func (f *chatRepoFake) UpdateMemberLastSeenTime(context.Context, string, string, int64) error {
	return nil
}

func (f *chatRepoFake) ListMyGroups(context.Context, string) ([]repo.ChatGroupListItem, error) {
	return nil, nil
}

func (f *chatRepoFake) ListDiscoverGroups(context.Context, string, string, string, string, int) ([]repo.ChatGroupDiscoverItem, string, bool, error) {
	return nil, "", false, nil
}

func (f *chatRepoFake) SearchGroupsByName(context.Context, string, int) ([]repo.ChatGroupDiscoverItem, error) {
	return nil, nil
}

func (f *chatRepoFake) GetGroupProfile(context.Context, string, int) (repo.ChatGroupProfileItem, []repo.ChatGroupMemberProfile, error) {
	return repo.ChatGroupProfileItem{}, nil, nil
}

func (f *chatRepoFake) ListMembersWithProfile(context.Context, string, int, int) ([]repo.ChatGroupMemberProfile, bool, error) {
	return nil, false, nil
}

func (f *chatRepoFake) SaveMessage(context.Context, string, string, string, string, string, int64) error {
	return nil
}

func (f *chatRepoFake) SaveDirectMessage(_ context.Context, messageID, fromUserID, toUserID, msgType, content string, now int64) error {
	f.savedDirectMessage = savedDirectMessage{
		messageID:  messageID,
		fromUserID: fromUserID,
		toUserID:   toUserID,
		msgType:    msgType,
		content:    content,
		now:        now,
	}
	return f.saveDirectMessageErr
}

func (f *chatRepoFake) ListMessagesLatest(context.Context, string, int) ([]repo.ChatGroupMessageItem, error) {
	return nil, nil
}

func (f *chatRepoFake) ListMessagesBefore(context.Context, string, int64, string, int) ([]repo.ChatGroupMessageItem, error) {
	return nil, nil
}

func (f *chatRepoFake) ListMessagesAfter(context.Context, string, int64, string, int) ([]repo.ChatGroupMessageItem, error) {
	return nil, nil
}

func (f *chatRepoFake) GetMemberRole(context.Context, string, string) (string, error) {
	if f.getMemberRoleErr != nil {
		return "", f.getMemberRoleErr
	}
	return "member", nil
}

func TestChatLogicListMessagesRejectsInvalidCursor(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{}, event.NoopPublisher{})

	_, err := l.ListMessages(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"}, dto.ChatGroupMessageListReq{
		Before: "bad-cursor",
	})

	if !errors.Is(err, logic.ErrInvalidCursor) {
		t.Fatalf("ListMessages() error = %v, want %v", err, logic.ErrInvalidCursor)
	}
}

func TestChatLogicListMessagesRejectsBeforeAndAfter(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{}, event.NoopPublisher{})

	_, err := l.ListMessages(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"}, dto.ChatGroupMessageListReq{
		Before: "123|00000000-0000-0000-0000-000000000001",
		After:  "123|00000000-0000-0000-0000-000000000002",
	})

	if !errors.Is(err, logic.ErrParamsType) {
		t.Fatalf("ListMessages() error = %v, want %v", err, logic.ErrParamsType)
	}
}

func TestChatLogicSaveMessageRejectsInvalidType(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{}, event.NoopPublisher{})

	err := l.SaveMessage(t.Context(), "user-1", "group-1", "message-1", "video", "hello", 1)

	if !errors.Is(err, logic.ErrInvalidMessageType) {
		t.Fatalf("SaveMessage() error = %v, want %v", err, logic.ErrInvalidMessageType)
	}
}

func TestChatLogicSaveMessageMapsNotMember(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{getMemberRoleErr: repo.ErrNotMember}, allowAllLimiter{}, event.NoopPublisher{})

	err := l.SaveMessage(t.Context(), "user-1", "group-1", "message-1", constant.MessageTypeText, "hello", 1)

	if !errors.Is(err, logic.ErrNotMember) {
		t.Fatalf("SaveMessage() error = %v, want %v", err, logic.ErrNotMember)
	}
}

func TestChatLogicHandleDirectMessageStoresAndReturnsMessage(t *testing.T) {
	chatRepo := &chatRepoFake{}
	publisher := &publisherFake{}
	l := logic.NewChatLogic(chatRepo, allowAllLimiter{}, publisher)

	got, err := l.HandleDirectMessage(t.Context(), "sender-id", "receiver-id", []byte(" hello\nthere "))

	if err != nil {
		t.Fatalf("HandleDirectMessage() error = %v", err)
	}
	if got.RecipientID != "receiver-id" {
		t.Fatalf("HandleDirectMessage() recipient = %q, want receiver-id", got.RecipientID)
	}
	if string(got.Message) != "hello there" {
		t.Fatalf("HandleDirectMessage() message = %q, want hello there", got.Message)
	}
	if _, err := uuid.Parse(chatRepo.savedDirectMessage.messageID); err != nil {
		t.Fatalf("saved message id = %q, want uuid: %v", chatRepo.savedDirectMessage.messageID, err)
	}
	if chatRepo.savedDirectMessage.fromUserID != "sender-id" {
		t.Fatalf("saved from user = %q, want sender-id", chatRepo.savedDirectMessage.fromUserID)
	}
	if chatRepo.savedDirectMessage.toUserID != "receiver-id" {
		t.Fatalf("saved to user = %q, want receiver-id", chatRepo.savedDirectMessage.toUserID)
	}
	if chatRepo.savedDirectMessage.msgType != constant.MessageTypeText {
		t.Fatalf("saved message type = %q, want %q", chatRepo.savedDirectMessage.msgType, constant.MessageTypeText)
	}
	if chatRepo.savedDirectMessage.content != "hello there" {
		t.Fatalf("saved content = %q, want hello there", chatRepo.savedDirectMessage.content)
	}
	if chatRepo.savedDirectMessage.now <= 0 {
		t.Fatalf("saved time = %d, want positive value", chatRepo.savedDirectMessage.now)
	}
	if publisher.direct.EventID != event.DirectEventID(chatRepo.savedDirectMessage.messageID) {
		t.Fatalf("published event id = %q, want %q", publisher.direct.EventID, event.DirectEventID(chatRepo.savedDirectMessage.messageID))
	}
	if publisher.direct.ToUserID != "receiver-id" || publisher.direct.Content != "hello there" {
		t.Fatalf("published direct message = %+v, want receiver/content", publisher.direct)
	}
}

func TestChatLogicHandleDirectMessageRejectsInvalidMessage(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		partnerID string
		message   []byte
	}{
		{name: "missing sender", userID: "", partnerID: "receiver-id", message: []byte("hello")},
		{name: "missing receiver", userID: "sender-id", partnerID: "", message: []byte("hello")},
		{name: "same user", userID: "sender-id", partnerID: "sender-id", message: []byte("hello")},
		{name: "empty message", userID: "sender-id", partnerID: "receiver-id", message: []byte(" \n ")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &chatRepoFake{}
			l := logic.NewChatLogic(chatRepo, allowAllLimiter{}, event.NoopPublisher{})

			got, err := l.HandleDirectMessage(t.Context(), tt.userID, tt.partnerID, tt.message)

			if !errors.Is(err, logic.ErrParamsType) {
				t.Fatalf("HandleDirectMessage() error = %v, want %v", err, logic.ErrParamsType)
			}
			if got.RecipientID != "" || len(got.Message) != 0 {
				t.Fatalf("HandleDirectMessage() result = %+v, want empty result", got)
			}
			if chatRepo.savedDirectMessage != (savedDirectMessage{}) {
				t.Fatalf("saved direct message = %+v, want none", chatRepo.savedDirectMessage)
			}
		})
	}
}

func TestChatLogicHandleDirectMessageMapsRepoError(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{saveDirectMessageErr: repo.ErrDefault}, allowAllLimiter{}, event.NoopPublisher{})

	got, err := l.HandleDirectMessage(t.Context(), "sender-id", "receiver-id", []byte("hello"))

	if !errors.Is(err, logic.ErrDefault) {
		t.Fatalf("HandleDirectMessage() error = %v, want %v", err, logic.ErrDefault)
	}
	if got.RecipientID != "" || len(got.Message) != 0 {
		t.Fatalf("HandleDirectMessage() result = %+v, want empty result", got)
	}
}

func TestChatLogicHandleGroupMessagePublishesEvent(t *testing.T) {
	publisher := &publisherFake{}
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{}, publisher)
	groupID := uuid.NewString()
	messageID := uuid.NewString()

	in := []byte(`{"op":"send","group_id":"` + groupID + `","message_id":"` + messageID + `","type":"text","content":"hello"}`)
	got := l.HandleGroupMessage(t.Context(), "user-1", in)

	if got.Ack == nil || !got.Ack.Ok {
		t.Fatalf("HandleGroupMessage() ack = %+v, want ok", got.Ack)
	}
	if publisher.group.EventID != event.GroupEventID(groupID, messageID) {
		t.Fatalf("published event id = %q, want %q", publisher.group.EventID, event.GroupEventID(groupID, messageID))
	}
	if publisher.group.GroupID != groupID || publisher.group.MessageID != messageID || publisher.group.Content != "hello" {
		t.Fatalf("published group message = %+v, want group/message/content", publisher.group)
	}
	if publisher.group.Payload == "" {
		t.Fatal("published group payload is empty")
	}
}

func TestChatLogicMapsRepoErrors(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{name: "group full", repoErr: repo.ErrGroupFull, wantErr: logic.ErrGroupFull},
		{name: "permission denied", repoErr: repo.ErrPermissionDenied, wantErr: logic.ErrPermissionDenied},
		{name: "unexpected repo error", repoErr: errors.New("database failed"), wantErr: logic.ErrDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logic.NewChatLogic(&chatRepoFake{joinGroupErr: tt.repoErr}, allowAllLimiter{}, event.NoopPublisher{})

			err := l.JoinGroup(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("JoinGroup() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
