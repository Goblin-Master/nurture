package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/event"
	"nurture/internal/chat/repo"
	"strings"
	"time"
)

func (l *ChatLogic) SaveMessage(ctx context.Context, userID string, groupID string, messageID string, msgType string, content string, now int64) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(messageID) == "" {
		return ErrParamsType
	}
	if !IsChatMessageType(msgType) {
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
	out, err := json.Marshal(GroupOutMessage{
		Op:      "new_message",
		GroupID: groupID,
		Message: GroupMessageBody{
			MessageID:  messageID,
			FromUserID: userID,
			Type:       msgType,
			Content:    content,
			Ctime:      now,
		},
	})
	if err != nil {
		return ErrDefault
	}
	payload, err := json.Marshal(event.GroupMessage{
		EventID:    event.GroupEventID(groupID, messageID),
		MessageID:  messageID,
		GroupID:    groupID,
		FromUserID: userID,
		Type:       msgType,
		Content:    content,
		Ctime:      now,
		Payload:    string(out),
	})
	if err != nil {
		return ErrDefault
	}
	if _, err := l.chatRepo.SaveMessage(ctx, groupID, messageID, userID, msgType, content, now, repo.ChatOutboxEvent{
		EventID:    event.GroupEventID(groupID, messageID),
		RoutingKey: event.RoutingKeyGroup,
		Payload:    string(payload),
		Ctime:      now,
	}); err != nil {
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
