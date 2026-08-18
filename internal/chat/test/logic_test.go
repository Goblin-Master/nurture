package test

import (
	"context"
	"errors"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/dto"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/repo"
	"testing"
	"time"
)

type allowAllLimiter struct{}

func (allowAllLimiter) Allow(_ context.Context, _ string, limit int64, _ time.Duration) (bool, int64, error) {
	return true, limit, nil
}

type chatRepoFake struct {
	joinGroupErr     error
	getMemberRoleErr error
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
