package logic

import (
	"context"
	"nurture/internal/chat/dto"
	"nurture/internal/chat/repo"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error)
}

type IChatLogic interface {
	CreateGroup(ctx context.Context, userID string, req dto.CreateChatGroupReq) (dto.CreateChatGroupResp, error)
	JoinGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error
	LeaveGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error
	TransferOwner(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupTransferReq) error
	DissolveGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error

	ListMyGroups(ctx context.Context, userID string) (dto.ListMyChatGroupsResp, error)
	DiscoverGroups(ctx context.Context, userID string, req dto.ChatGroupDiscoverReq) (dto.ChatGroupDiscoverResp, error)
	SearchGroups(ctx context.Context, userID string, req dto.ChatGroupSearchReq) (dto.ChatGroupSearchResp, error)
	GetGroupProfile(ctx context.Context, userID string, uri dto.ChatGroupIDUri) (dto.ChatGroupProfileResp, error)
	ListMembers(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupMemberListReq) (dto.ChatGroupMemberListResp, error)
	ListMessages(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupMessageListReq) (dto.ChatGroupMessageListResp, error)

	SaveMessage(ctx context.Context, userID string, groupID string, messageID string, msgType string, content string, now int64) error
	MarkGroupSeen(ctx context.Context, userID string, groupID string, now int64) error
	CheckMember(ctx context.Context, userID string, groupID string) error
	HandleDirectMessage(ctx context.Context, userID string, partnerID string, message []byte) (DirectMessageResult, error)
	HandleGroupMessage(ctx context.Context, userID string, message []byte) GroupMessageResult
}

type ChatLogic struct {
	chatRepo repo.IChatRepo
	limiter  RateLimiter
}

func NewChatLogic(chatRepo repo.IChatRepo, limiter RateLimiter) *ChatLogic {
	return &ChatLogic{
		chatRepo: chatRepo,
		limiter:  limiter,
	}
}

var _ IChatLogic = (*ChatLogic)(nil)
