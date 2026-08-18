package chat

import (
	"context"
	"fmt"
	"nurture/internal/dto"
	"nurture/internal/global"
	"strings"
	"time"
)

func (l *Logic) ListMessages(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupMessageListReq) (dto.ChatGroupMessageListResp, error) {
	var resp dto.ChatGroupMessageListResp
	if strings.TrimSpace(uri.GroupID) == "" {
		return resp, ErrParamsType
	}
	if _, err := l.chatRepo.GetMemberRole(ctx, uri.GroupID, userID); err != nil {
		return resp, mapRepoErr(err)
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if req.Before != "" && req.After != "" {
		return resp, ErrParamsType
	}
	if req.After != "" {
		ctime, mid, err := parseCursor(req.After)
		if err != nil {
			return resp, err
		}
		rows, err := l.chatRepo.ListMessagesAfter(ctx, uri.GroupID, ctime, mid, limit)
		if err != nil {
			return resp, mapRepoErr(err)
		}
		resp.Items = make([]dto.ChatGroupMessageItem, 0, len(rows))
		for _, v := range rows {
			resp.Items = append(resp.Items, dto.ChatGroupMessageItem{
				MessageID:  v.MessageID,
				GroupID:    v.GroupID,
				FromUserID: v.FromUserID,
				Type:       v.Type,
				Content:    v.Content,
				Ctime:      v.Ctime,
			})
		}
		if len(resp.Items) > 0 {
			last := resp.Items[len(resp.Items)-1]
			resp.NextAfter = buildCursor(last.Ctime, last.MessageID)
		}
		resp.HasMore = len(resp.Items) >= limit
		return resp, nil
	}
	if req.Before != "" {
		ctime, mid, err := parseCursor(req.Before)
		if err != nil {
			return resp, err
		}
		rows, err := l.chatRepo.ListMessagesBefore(ctx, uri.GroupID, ctime, mid, limit)
		if err != nil {
			return resp, mapRepoErr(err)
		}
		resp.Items = make([]dto.ChatGroupMessageItem, 0, len(rows))
		for _, v := range rows {
			resp.Items = append(resp.Items, dto.ChatGroupMessageItem{
				MessageID:  v.MessageID,
				GroupID:    v.GroupID,
				FromUserID: v.FromUserID,
				Type:       v.Type,
				Content:    v.Content,
				Ctime:      v.Ctime,
			})
		}
		if len(resp.Items) > 0 {
			last := resp.Items[len(resp.Items)-1]
			resp.NextBefore = buildCursor(last.Ctime, last.MessageID)
		}
		resp.HasMore = len(resp.Items) >= limit
		return resp, nil
	}
	rows, err := l.chatRepo.ListMessagesLatest(ctx, uri.GroupID, limit)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Items = make([]dto.ChatGroupMessageItem, 0, len(rows))
	for _, v := range rows {
		resp.Items = append(resp.Items, dto.ChatGroupMessageItem{
			MessageID:  v.MessageID,
			GroupID:    v.GroupID,
			FromUserID: v.FromUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	if len(resp.Items) > 0 {
		last := resp.Items[len(resp.Items)-1]
		resp.NextBefore = buildCursor(last.Ctime, last.MessageID)
	}
	resp.HasMore = len(resp.Items) >= limit
	return resp, nil
}

func (l *Logic) SaveMessage(ctx context.Context, userID string, groupID string, messageID string, msgType string, content string, now int64) error {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(messageID) == "" {
		return ErrParamsType
	}
	if msgType != "text" && msgType != "image" && msgType != "system" {
		return ErrInvalidMessageType
	}
	if content == "" {
		return ErrParamsType
	}
	if global.RDB != nil {
		msgLimiter.SetRedis(global.RDB)
	}
	if ok, _, _ := msgLimiter.Allow(ctx, fmt.Sprintf("rl:chat:send:user:%s", userID), 10, time.Second); !ok {
		return ErrTooManyRequests
	}
	if ok, _, _ := msgLimiter.Allow(ctx, fmt.Sprintf("rl:chat:send:group:%s:%s", groupID, userID), 5, time.Second); !ok {
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

func (l *Logic) MarkGroupSeen(ctx context.Context, userID string, groupID string, now int64) error {
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

func (l *Logic) CheckMember(ctx context.Context, userID string, groupID string) error {
	if strings.TrimSpace(groupID) == "" {
		return ErrParamsType
	}
	if _, err := l.chatRepo.GetMemberRole(ctx, groupID, userID); err != nil {
		return mapRepoErr(err)
	}
	return nil
}
