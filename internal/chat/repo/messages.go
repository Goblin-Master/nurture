package repo

import (
	"context"
	"errors"
	"nurture/internal/chat/repo/dao"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *ChatRepo) SaveMessage(ctx context.Context, groupID, messageID, fromUserID, msgType, content string, now int64) error {
	var gid, mid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := mid.Scan(messageID); err != nil {
		return ErrParamsType
	}
	if err := uid.Scan(fromUserID); err != nil {
		return ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	aff, err := r.dao.CreateChatGroupMessage(ctx, dao.CreateChatGroupMessageParams{
		MessageID:  mid,
		GroupID:    gid,
		FromUserID: uid,
		Type:       msgType,
		Content:    content,
		Ctime:      now,
	})
	if err != nil {
		r.logError(err)
		return ErrDefault
	}
	if aff == 0 {
		return nil
	}
	return nil
}

func (r *ChatRepo) ListMessagesLatest(ctx context.Context, groupID string, limit int) ([]ChatGroupMessageItem, error) {
	var gid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatGroupMessagesLatest(ctx, dao.ListChatGroupMessagesLatestParams{
		GroupID: gid,
		Limit:   int32(limit),
	})
	if err != nil {
		r.logError(err)
		return nil, ErrDefault
	}
	items := make([]ChatGroupMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatGroupMessageItem{
			MessageID:  v.MessageID,
			GroupID:    v.GroupID,
			FromUserID: v.FromUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) ListMessagesBefore(ctx context.Context, groupID string, beforeCtime int64, beforeMessageID string, limit int) ([]ChatGroupMessageItem, error) {
	var gid, mid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return nil, ErrParamsType
	}
	if err := mid.Scan(beforeMessageID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatGroupMessagesBefore(ctx, dao.ListChatGroupMessagesBeforeParams{
		GroupID:   gid,
		Ctime:     beforeCtime,
		MessageID: mid,
		Limit:     int32(limit),
	})
	if err != nil {
		r.logError(err)
		return nil, ErrDefault
	}
	items := make([]ChatGroupMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatGroupMessageItem{
			MessageID:  v.MessageID,
			GroupID:    v.GroupID,
			FromUserID: v.FromUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) ListMessagesAfter(ctx context.Context, groupID string, afterCtime int64, afterMessageID string, limit int) ([]ChatGroupMessageItem, error) {
	var gid, mid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return nil, ErrParamsType
	}
	if err := mid.Scan(afterMessageID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatGroupMessagesAfter(ctx, dao.ListChatGroupMessagesAfterParams{
		GroupID:   gid,
		Ctime:     afterCtime,
		MessageID: mid,
		Limit:     int32(limit),
	})
	if err != nil {
		r.logError(err)
		return nil, ErrDefault
	}
	items := make([]ChatGroupMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatGroupMessageItem{
			MessageID:  v.MessageID,
			GroupID:    v.GroupID,
			FromUserID: v.FromUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) GetMemberRole(ctx context.Context, groupID, userID string) (string, error) {
	var gid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return "", ErrParamsType
	}
	if err := uid.Scan(userID); err != nil {
		return "", ErrParamsType
	}
	role, err := r.dao.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{
		GroupID: gid,
		UserID:  uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotMember
		}
		r.logError(err)
		return "", ErrDefault
	}
	return role, nil
}
