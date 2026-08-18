package logic

import (
	"context"
	"fmt"
	"nurture/internal/chat/constant"
	"strings"
	"time"
)

func (l *ChatLogic) SaveMessage(ctx context.Context, userID string, groupID string, messageID string, msgType string, content string, now int64) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(messageID) == "" {
		return ErrParamsType
	}
	if msgType != constant.MessageTypeText && msgType != constant.MessageTypeImage && msgType != constant.MessageTypeSystem {
		return ErrInvalidMessageType
	}
	if content == "" {
		return ErrParamsType
	}
	if l.limiter == nil {
		return ErrDefault
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendUserKey, userID), constant.RateLimitWSSendUserLimit, constant.RateLimitWSSendWindow); !ok {
		return ErrTooManyRequests
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendGroupKey, groupID, userID), constant.RateLimitWSSendGroupLimit, constant.RateLimitWSSendWindow); !ok {
		return ErrTooManyRequests
	}
	if _, err := l.chatRepo.GetMemberRole(ctx, groupID, userID); err != nil {
		return mapRepoErr(err)
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if err := l.chatRepo.SaveMessage(ctx, groupID, messageID, userID, msgType, content, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) MarkGroupSeen(ctx context.Context, userID string, groupID string, now int64) error {
	if strings.TrimSpace(groupID) == "" {
		return ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if err := l.chatRepo.UpdateMemberLastSeenTime(ctx, groupID, userID, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) CheckMember(ctx context.Context, userID string, groupID string) error {
	if strings.TrimSpace(groupID) == "" {
		return ErrParamsType
	}
	if _, err := l.chatRepo.GetMemberRole(ctx, groupID, userID); err != nil {
		return mapRepoErr(err)
	}
	return nil
}
