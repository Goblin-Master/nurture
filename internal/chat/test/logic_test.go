package test

import (
	"context"
	"encoding/json"
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

type chatRepoFake struct {
	joinGroupErr     error
	getMemberRoleErr error

	saveDirectMessageErr error
	saveDirectCreated    bool
	savedDirectMessage   savedDirectMessage

	saveGroupCreated  bool
	savedGroupOutbox  repo.ChatOutboxEvent
	directMessageRows []repo.ChatDirectMessageItem
}

type savedDirectMessage struct {
	messageID  string
	fromUserID string
	toUserID   string
	msgType    string
	content    string
	now        int64
	outbox     repo.ChatOutboxEvent
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

func (f *chatRepoFake) SaveMessage(_ context.Context, groupID, messageID, fromUserID, msgType, content string, now int64, outbox repo.ChatOutboxEvent) (bool, error) {
	f.savedGroupOutbox = outbox
	return f.saveGroupCreated, nil
}

func (f *chatRepoFake) SaveDirectMessage(_ context.Context, messageID, fromUserID, toUserID, msgType, content string, now int64, outbox repo.ChatOutboxEvent) (bool, error) {
	f.savedDirectMessage = savedDirectMessage{
		messageID:  messageID,
		fromUserID: fromUserID,
		toUserID:   toUserID,
		msgType:    msgType,
		content:    content,
		now:        now,
		outbox:     outbox,
	}
	return f.saveDirectCreated, f.saveDirectMessageErr
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

func (f *chatRepoFake) ListDirectMessagesLatest(context.Context, string, string, int) ([]repo.ChatDirectMessageItem, error) {
	return f.directMessageRows, nil
}

func (f *chatRepoFake) ListDirectMessagesBefore(context.Context, string, string, int64, string, int) ([]repo.ChatDirectMessageItem, error) {
	return f.directMessageRows, nil
}

func (f *chatRepoFake) ListDirectMessagesAfter(context.Context, string, string, int64, string, int) ([]repo.ChatDirectMessageItem, error) {
	return f.directMessageRows, nil
}

func (f *chatRepoFake) MarkDirectSeen(context.Context, string, string, int64, int64) error {
	return nil
}

func (f *chatRepoFake) GetMemberRole(context.Context, string, string) (string, error) {
	if f.getMemberRoleErr != nil {
		return "", f.getMemberRoleErr
	}
	return "member", nil
}

func TestChatLogicListMessagesRejectsInvalidCursor(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{})

	_, err := l.ListMessages(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"}, dto.ChatGroupMessageListReq{
		Before: "bad-cursor",
	})

	if !errors.Is(err, logic.ErrInvalidCursor) {
		t.Fatalf("ListMessages() error = %v, want %v", err, logic.ErrInvalidCursor)
	}
}

func TestChatLogicListMessagesRejectsBeforeAndAfter(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{})

	_, err := l.ListMessages(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"}, dto.ChatGroupMessageListReq{
		Before: "123|00000000-0000-0000-0000-000000000001",
		After:  "123|00000000-0000-0000-0000-000000000002",
	})

	if !errors.Is(err, logic.ErrParamsType) {
		t.Fatalf("ListMessages() error = %v, want %v", err, logic.ErrParamsType)
	}
}

func TestChatLogicSaveMessageRejectsInvalidType(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{}, allowAllLimiter{})

	err := l.SaveMessage(t.Context(), "user-1", "group-1", "message-1", "video", "hello", 1)

	if !errors.Is(err, logic.ErrInvalidMessageType) {
		t.Fatalf("SaveMessage() error = %v, want %v", err, logic.ErrInvalidMessageType)
	}
}

func TestChatLogicSaveMessageMapsNotMember(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{getMemberRoleErr: repo.ErrNotMember}, allowAllLimiter{})

	err := l.SaveMessage(t.Context(), "user-1", "group-1", "message-1", constant.MessageTypeText, "hello", 1)

	if !errors.Is(err, logic.ErrNotMember) {
		t.Fatalf("SaveMessage() error = %v, want %v", err, logic.ErrNotMember)
	}
}

func TestChatLogicHandleDirectMessageStoresOutboxAndReturnsAck(t *testing.T) {
	messageID := uuid.NewString()
	chatRepo := &chatRepoFake{saveDirectCreated: true}
	l := logic.NewChatLogic(chatRepo, allowAllLimiter{})

	in := []byte(`{"op":"send","message_id":"` + messageID + `","type":"text","content":"hello"}`)
	got, err := l.HandleDirectMessage(t.Context(), "sender-id", "receiver-id", in)

	if err != nil {
		t.Fatalf("HandleDirectMessage() error = %v", err)
	}
	if got.Ack == nil || !got.Ack.Ok {
		t.Fatalf("HandleDirectMessage() ack = %+v, want ok", got.Ack)
	}
	if got.Ack.MessageID != messageID {
		t.Fatalf("ack message id = %q, want %q", got.Ack.MessageID, messageID)
	}
	if chatRepo.savedDirectMessage.messageID != messageID {
		t.Fatalf("saved message id = %q, want %q", chatRepo.savedDirectMessage.messageID, messageID)
	}
	if chatRepo.savedDirectMessage.fromUserID != "sender-id" || chatRepo.savedDirectMessage.toUserID != "receiver-id" {
		t.Fatalf("saved users = %+v, want sender/receiver", chatRepo.savedDirectMessage)
	}
	if chatRepo.savedDirectMessage.msgType != constant.MessageTypeText {
		t.Fatalf("saved message type = %q, want %q", chatRepo.savedDirectMessage.msgType, constant.MessageTypeText)
	}
	if chatRepo.savedDirectMessage.content != "hello" {
		t.Fatalf("saved content = %q, want hello", chatRepo.savedDirectMessage.content)
	}
	if chatRepo.savedDirectMessage.now <= 0 {
		t.Fatalf("saved time = %d, want positive value", chatRepo.savedDirectMessage.now)
	}
	wantEventID := event.DirectEventID("sender-id", "receiver-id", messageID)
	if chatRepo.savedDirectMessage.outbox.EventID != wantEventID {
		t.Fatalf("outbox event id = %q, want %q", chatRepo.savedDirectMessage.outbox.EventID, wantEventID)
	}
	if chatRepo.savedDirectMessage.outbox.RoutingKey != event.RoutingKeyDirect {
		t.Fatalf("outbox routing key = %q, want %q", chatRepo.savedDirectMessage.outbox.RoutingKey, event.RoutingKeyDirect)
	}
	var out event.DirectMessage
	if err := json.Unmarshal([]byte(chatRepo.savedDirectMessage.outbox.Payload), &out); err != nil {
		t.Fatalf("outbox payload json error: %v", err)
	}
	if out.EventID != wantEventID || out.MessageID != messageID || out.ToUserID != "receiver-id" || out.Content != "hello" {
		t.Fatalf("outbox payload = %+v, want event/message/receiver/content", out)
	}
}

func TestChatLogicHandleDirectMessageRejectsInvalidMessage(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		partnerID string
		message   []byte
	}{
		{name: "missing sender", userID: "", partnerID: "receiver-id", message: []byte(`{"op":"send","message_id":"` + uuid.NewString() + `","type":"text","content":"hello"}`)},
		{name: "missing receiver", userID: "sender-id", partnerID: "", message: []byte(`{"op":"send","message_id":"` + uuid.NewString() + `","type":"text","content":"hello"}`)},
		{name: "same user", userID: "sender-id", partnerID: "sender-id", message: []byte(`{"op":"send","message_id":"` + uuid.NewString() + `","type":"text","content":"hello"}`)},
		{name: "empty message", userID: "sender-id", partnerID: "receiver-id", message: []byte(`{"op":"send","message_id":"` + uuid.NewString() + `","type":"text","content":" "}`)},
		{name: "plain text", userID: "sender-id", partnerID: "receiver-id", message: []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRepo := &chatRepoFake{}
			l := logic.NewChatLogic(chatRepo, allowAllLimiter{})

			got, err := l.HandleDirectMessage(t.Context(), tt.userID, tt.partnerID, tt.message)

			if err != nil {
				t.Fatalf("HandleDirectMessage() error = %v, want %v", err, logic.ErrParamsType)
			}
			if got.Ack == nil || got.Ack.Ok {
				t.Fatalf("HandleDirectMessage() ack = %+v, want failed ack", got.Ack)
			}
			if chatRepo.savedDirectMessage != (savedDirectMessage{}) {
				t.Fatalf("saved direct message = %+v, want none", chatRepo.savedDirectMessage)
			}
		})
	}
}

func TestChatLogicHandleDirectMessageMapsRepoError(t *testing.T) {
	l := logic.NewChatLogic(&chatRepoFake{saveDirectMessageErr: repo.ErrDefault}, allowAllLimiter{})

	got, err := l.HandleDirectMessage(t.Context(), "sender-id", "receiver-id", []byte(`{"op":"send","message_id":"`+uuid.NewString()+`","type":"text","content":"hello"}`))

	if err != nil {
		t.Fatalf("HandleDirectMessage() error = %v, want nil with failed ack", err)
	}
	if got.Ack == nil || got.Ack.Ok || got.Ack.Error != logic.ErrDefault.Error() {
		t.Fatalf("HandleDirectMessage() ack = %+v, want default error ack", got.Ack)
	}
}

func TestChatLogicHandleGroupMessageStoresOutbox(t *testing.T) {
	chatRepo := &chatRepoFake{saveGroupCreated: true}
	l := logic.NewChatLogic(chatRepo, allowAllLimiter{})
	groupID := uuid.NewString()
	messageID := uuid.NewString()

	in := []byte(`{"op":"send","group_id":"` + groupID + `","message_id":"` + messageID + `","type":"text","content":"hello"}`)
	got := l.HandleGroupMessage(t.Context(), "user-1", in)

	if got.Ack == nil || !got.Ack.Ok {
		t.Fatalf("HandleGroupMessage() ack = %+v, want ok", got.Ack)
	}
	if chatRepo.savedGroupOutbox.EventID != event.GroupEventID(groupID, messageID) {
		t.Fatalf("outbox event id = %q, want %q", chatRepo.savedGroupOutbox.EventID, event.GroupEventID(groupID, messageID))
	}
	if chatRepo.savedGroupOutbox.RoutingKey != event.RoutingKeyGroup {
		t.Fatalf("outbox routing key = %q, want %q", chatRepo.savedGroupOutbox.RoutingKey, event.RoutingKeyGroup)
	}
	var out event.GroupMessage
	if err := json.Unmarshal([]byte(chatRepo.savedGroupOutbox.Payload), &out); err != nil {
		t.Fatalf("outbox payload json error: %v", err)
	}
	if out.GroupID != groupID || out.MessageID != messageID || out.Content != "hello" || out.Payload == "" {
		t.Fatalf("outbox payload = %+v, want group/message/content/payload", out)
	}
}

func TestChatLogicListDirectMessagesMapsRows(t *testing.T) {
	partnerID := uuid.NewString()
	messageID := uuid.NewString()
	l := logic.NewChatLogic(&chatRepoFake{directMessageRows: []repo.ChatDirectMessageItem{{
		MessageID:  messageID,
		FromUserID: "user-1",
		ToUserID:   partnerID,
		Type:       constant.MessageTypeText,
		Content:    "hello",
		Ctime:      123,
	}}}, allowAllLimiter{})

	got, err := l.ListDirectMessages(t.Context(), "user-1", dto.ChatDirectMessageUserUri{UserID: partnerID}, dto.ChatDirectMessageListReq{})

	if err != nil {
		t.Fatalf("ListDirectMessages() error = %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("ListDirectMessages() items = %d, want 1", len(got.Items))
	}
	if got.Items[0].MessageID != messageID || got.Items[0].ToUserID != partnerID {
		t.Fatalf("ListDirectMessages() item = %+v, want message/partner", got.Items[0])
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
			l := logic.NewChatLogic(&chatRepoFake{joinGroupErr: tt.repoErr}, allowAllLimiter{})

			err := l.JoinGroup(t.Context(), "user-1", dto.ChatGroupIDUri{GroupID: "group-1"})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("JoinGroup() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
