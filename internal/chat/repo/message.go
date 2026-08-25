package repo

import (
	"context"
	"errors"
	"nurture/internal/chat/repo/dao"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *ChatRepo) SaveMessage(ctx context.Context, groupID, messageID, fromUserID, msgType, content string, now int64, outbox ChatOutboxEvent) (bool, error) {
	var gid, mid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return false, ErrParamsType
	}
	if err := mid.Scan(messageID); err != nil {
		return false, ErrParamsType
	}
	if err := uid.Scan(fromUserID); err != nil {
		return false, ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	defer tx.Rollback(ctx)

	qtx := r.dao.WithTx(tx)
	aff, err := qtx.CreateChatGroupMessage(ctx, dao.CreateChatGroupMessageParams{
		MessageID:  mid,
		GroupID:    gid,
		FromUserID: uid,
		Type:       msgType,
		Content:    content,
		Ctime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	if aff == 0 {
		return false, nil
	}
	if err := r.createOutbox(ctx, qtx, outbox, now); err != nil {
		return false, err
	}
	return true, r.commit(ctx, tx)
}

func (r *ChatRepo) SaveDirectMessage(ctx context.Context, messageID, fromUserID, toUserID, msgType, content string, now int64, outbox ChatOutboxEvent) (bool, error) {
	var mid, fromUID, toUID pgtype.UUID
	if err := mid.Scan(messageID); err != nil {
		return false, ErrParamsType
	}
	if err := fromUID.Scan(fromUserID); err != nil {
		return false, ErrParamsType
	}
	if err := toUID.Scan(toUserID); err != nil {
		return false, ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	defer tx.Rollback(ctx)

	qtx := r.dao.WithTx(tx)
	aff, err := qtx.CreateChatDirectMessage(ctx, dao.CreateChatDirectMessageParams{
		MessageID:  mid,
		FromUserID: fromUID,
		ToUserID:   toUID,
		Type:       msgType,
		Content:    content,
		Ctime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	if aff == 0 {
		return false, nil
	}
	if err := r.createOutbox(ctx, qtx, outbox, now); err != nil {
		return false, err
	}
	return true, r.commit(ctx, tx)
}

func (r *ChatRepo) createOutbox(ctx context.Context, q *dao.Queries, outbox ChatOutboxEvent, now int64) error {
	if outbox.EventID == "" || outbox.RoutingKey == "" || outbox.Payload == "" {
		return ErrParamsType
	}
	if outbox.Ctime <= 0 {
		outbox.Ctime = now
	}
	aff, err := q.CreateChatEventOutbox(ctx, dao.CreateChatEventOutboxParams{
		EventID:    outbox.EventID,
		RoutingKey: outbox.RoutingKey,
		Payload:    outbox.Payload,
		Ctime:      outbox.Ctime,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrDefault
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
		r.log.Error(err)
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
		r.log.Error(err)
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
		r.log.Error(err)
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

func (r *ChatRepo) ListDirectMessagesLatest(ctx context.Context, userID, partnerID string, limit int) ([]ChatDirectMessageItem, error) {
	var uid, pid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, ErrParamsType
	}
	if err := pid.Scan(partnerID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatDirectMessagesLatest(ctx, dao.ListChatDirectMessagesLatestParams{
		FromUserID: uid,
		ToUserID:   pid,
		Limit:      int32(limit),
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatDirectMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatDirectMessageItem{
			MessageID:  v.MessageID,
			FromUserID: v.FromUserID,
			ToUserID:   v.ToUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) ListDirectMessagesBefore(ctx context.Context, userID, partnerID string, beforeCtime int64, beforeMessageID string, limit int) ([]ChatDirectMessageItem, error) {
	var uid, pid, mid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, ErrParamsType
	}
	if err := pid.Scan(partnerID); err != nil {
		return nil, ErrParamsType
	}
	if err := mid.Scan(beforeMessageID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatDirectMessagesBefore(ctx, dao.ListChatDirectMessagesBeforeParams{
		FromUserID: uid,
		ToUserID:   pid,
		Ctime:      beforeCtime,
		MessageID:  mid,
		Limit:      int32(limit),
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatDirectMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatDirectMessageItem{
			MessageID:  v.MessageID,
			FromUserID: v.FromUserID,
			ToUserID:   v.ToUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) ListDirectMessagesAfter(ctx context.Context, userID, partnerID string, afterCtime int64, afterMessageID string, limit int) ([]ChatDirectMessageItem, error) {
	var uid, pid, mid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, ErrParamsType
	}
	if err := pid.Scan(partnerID); err != nil {
		return nil, ErrParamsType
	}
	if err := mid.Scan(afterMessageID); err != nil {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ListChatDirectMessagesAfter(ctx, dao.ListChatDirectMessagesAfterParams{
		FromUserID: uid,
		ToUserID:   pid,
		Ctime:      afterCtime,
		MessageID:  mid,
		Limit:      int32(limit),
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatDirectMessageItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatDirectMessageItem{
			MessageID:  v.MessageID,
			FromUserID: v.FromUserID,
			ToUserID:   v.ToUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) MarkDirectSeen(ctx context.Context, userID, partnerID string, lastSeenTime int64, now int64) error {
	var uid, pid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}
	if err := pid.Scan(partnerID); err != nil {
		return ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if lastSeenTime <= 0 {
		lastSeenTime = now
	}
	if err := r.dao.UpsertChatDirectSeen(ctx, dao.UpsertChatDirectSeenParams{
		UserID:        uid,
		PartnerUserID: pid,
		LastSeenTime:  lastSeenTime,
		Ctime:         now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
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
		r.log.Error(err)
		return "", ErrDefault
	}
	return role, nil
}
