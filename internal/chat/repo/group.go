package repo

import (
	"context"
	"errors"
	"nurture/internal/chat/repo/dao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *ChatRepo) CreateGroup(ctx context.Context, groupID, ownerID, name, avatar, description string, memberLimit int32, now int64) error {
	var gid, oid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := oid.Scan(ownerID); err != nil {
		return ErrParamsType
	}

	if memberLimit <= 0 {
		return ErrParamsType
	}
	if err := r.dao.CreateChatGroup(ctx, dao.CreateChatGroupParams{
		GroupID:     gid,
		OwnerID:     oid,
		Name:        name,
		Avatar:      avatar,
		Description: description,
		MemberLimit: memberLimit,
		Ctime:       now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *ChatRepo) JoinGroup(ctx context.Context, groupID, userID string, now int64) error {
	var gid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGroupNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}

	if _, e := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: uid}); e == nil {
		return r.commit(ctx, tx)
	} else if !errors.Is(e, pgx.ErrNoRows) {
		r.log.Error(e)
		return ErrDefault
	}

	if g.MemberCount >= g.MemberLimit {
		return ErrGroupFull
	}

	if err := qtx.CreateChatGroupMember(ctx, dao.CreateChatGroupMemberParams{
		GroupID: gid,
		UserID:  uid,
		Role:    "member",
		Ctime:   now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if _, err := qtx.IncChatGroupMemberCount(ctx, dao.IncChatGroupMemberCountParams{
		GroupID: gid,
		Utime:   now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return r.commit(ctx, tx)
}

func (r *ChatRepo) LeaveGroup(ctx context.Context, groupID, userID string, now int64) error {
	var gid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	if _, err := qtx.LockChatGroupByID(ctx, gid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGroupNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	role, err := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotMember
		}
		r.log.Error(err)
		return ErrDefault
	}
	if role == "owner" {
		return ErrOwnerMustTransferFirst
	}
	aff, err := qtx.DeleteChatGroupMember(ctx, dao.DeleteChatGroupMemberParams{
		GroupID: gid,
		UserID:  uid,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrNotMember
	}
	if _, err := qtx.DecChatGroupMemberCount(ctx, dao.DecChatGroupMemberCountParams{
		GroupID: gid,
		Utime:   now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return r.commit(ctx, tx)
}

func (r *ChatRepo) TransferOwner(ctx context.Context, groupID, ownerID, targetUserID string, now int64) error {
	var gid, oid, tid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := oid.Scan(ownerID); err != nil {
		return ErrParamsType
	}
	if err := tid.Scan(targetUserID); err != nil {
		return ErrParamsType
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGroupNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	if g.OwnerID != oid {
		return ErrPermissionDenied
	}
	if _, err := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: tid}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotMember
		}
		r.log.Error(err)
		return ErrDefault
	}
	aff, err := qtx.TransferChatGroupOwner(ctx, dao.TransferChatGroupOwnerParams{
		GroupID:  gid,
		UserID:   tid,
		UserID_2: oid,
		Role:     "member",
		Utime:    now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff <= 0 {
		return ErrDefault
	}
	if _, err := qtx.UpdateChatGroupOwnerID(ctx, dao.UpdateChatGroupOwnerIDParams{
		GroupID: gid,
		OwnerID: tid,
		Utime:   now,
	}); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return r.commit(ctx, tx)
}

func (r *ChatRepo) DissolveGroup(ctx context.Context, groupID, ownerID string, now int64) error {
	var gid, oid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := oid.Scan(ownerID); err != nil {
		return ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrGroupNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	if g.OwnerID != oid {
		return ErrPermissionDenied
	}
	if _, err := qtx.DeleteChatGroupMessagesByGroupID(ctx, gid); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if _, err := qtx.DeleteChatGroupMembersByGroupID(ctx, gid); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	aff, err := qtx.DeleteChatGroupByID(ctx, gid)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrGroupNotExist
	}
	return r.commit(ctx, tx)
}

func (r *ChatRepo) UpdateMemberLastSeenTime(ctx context.Context, groupID, userID string, lastSeenTime int64) error {
	var gid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}
	if lastSeenTime <= 0 {
		lastSeenTime = time.Now().UnixMilli()
	}
	aff, err := r.dao.UpdateChatGroupMemberLastSeenTime(ctx, dao.UpdateChatGroupMemberLastSeenTimeParams{
		GroupID:      gid,
		UserID:       uid,
		LastSeenTime: lastSeenTime,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrNotMember
	}
	return nil
}

func (r *ChatRepo) ListMyGroups(ctx context.Context, userID string) ([]ChatGroupListItem, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, ErrParamsType
	}
	rows, err := r.dao.ListMyChatGroups(ctx, uid)
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatGroupListItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatGroupListItem{
			GroupID:               v.GroupID,
			Name:                  v.Name,
			Avatar:                v.Avatar,
			Description:           v.Description,
			MemberLimit:           v.MemberLimit,
			MemberCount:           v.MemberCount,
			Role:                  v.Role,
			UnreadCount:           v.UnreadCount,
			Ctime:                 v.Ctime,
			Utime:                 v.Utime,
			LastMessageFromUserID: toString(v.LastMessageFromUserID),
			LastMessageFromName:   v.LastMessageFromUsername,
			LastMessageType:       v.LastMessageType,
			LastMessageContent:    v.LastMessageContent,
			LastMessageTime:       v.LastMessageTime,
		})
	}
	return items, nil
}

func (r *ChatRepo) ListDiscoverGroups(ctx context.Context, userID, seed, cursorSortKey, cursorGroupID string, limit int) ([]ChatGroupDiscoverItem, string, bool, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, "", false, ErrParamsType
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	realLimit := int32(limit + 1)
	var rowsAny any
	if cursorSortKey == "" && cursorGroupID == "" {
		rows, err := r.dao.ListDiscoverChatGroupsFirst(ctx, dao.ListDiscoverChatGroupsFirstParams{
			UserID:  uid,
			Column2: seed,
			Limit:   realLimit,
		})
		if err != nil {
			r.log.Error(err)
			return nil, "", false, ErrDefault
		}
		rowsAny = rows
	} else {
		var gid pgtype.UUID
		if err := gid.Scan(cursorGroupID); err != nil {
			return nil, "", false, ErrParamsType
		}
		rows, err := r.dao.ListDiscoverChatGroupsAfter(ctx, dao.ListDiscoverChatGroupsAfterParams{
			UserID:  uid,
			Column2: seed,
			Column3: cursorSortKey,
			GroupID: gid,
			Limit:   realLimit,
		})
		if err != nil {
			r.log.Error(err)
			return nil, "", false, ErrDefault
		}
		rowsAny = rows
	}

	var items []ChatGroupDiscoverItem
	switch v := rowsAny.(type) {
	case []dao.ListDiscoverChatGroupsFirstRow:
		items = make([]ChatGroupDiscoverItem, 0, len(v))
		for _, r := range v {
			items = append(items, ChatGroupDiscoverItem{
				GroupID:     r.GroupID,
				Name:        r.Name,
				Avatar:      r.Avatar,
				MemberCount: r.MemberCount,
				MemberLimit: r.MemberLimit,
				SortKey:     r.SortKey,
			})
		}
	case []dao.ListDiscoverChatGroupsAfterRow:
		items = make([]ChatGroupDiscoverItem, 0, len(v))
		for _, r := range v {
			items = append(items, ChatGroupDiscoverItem{
				GroupID:     r.GroupID,
				Name:        r.Name,
				Avatar:      r.Avatar,
				MemberCount: r.MemberCount,
				MemberLimit: r.MemberLimit,
				SortKey:     r.SortKey,
			})
		}
	default:
		return nil, "", false, ErrDefault
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	nextCursor := ""
	if len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = last.SortKey + "|" + last.GroupID
	}
	return items, nextCursor, hasMore, nil
}

func (r *ChatRepo) SearchGroupsByName(ctx context.Context, keyword string, limit int) ([]ChatGroupDiscoverItem, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, ErrParamsType
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.SearchChatGroupsByName(ctx, dao.SearchChatGroupsByNameParams{
		Column1: pgtype.Text{String: keyword, Valid: true},
		Limit:   int32(limit),
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatGroupDiscoverItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatGroupDiscoverItem{
			GroupID:     v.GroupID,
			Name:        v.Name,
			Avatar:      v.Avatar,
			MemberCount: v.MemberCount,
			MemberLimit: v.MemberLimit,
		})
	}
	return items, nil
}

func (r *ChatRepo) GetGroupProfile(ctx context.Context, groupID string, memberPreviewLimit int) (ChatGroupProfileItem, []ChatGroupMemberProfile, error) {
	var gid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ChatGroupProfileItem{}, nil, ErrParamsType
	}
	if memberPreviewLimit <= 0 || memberPreviewLimit > 50 {
		memberPreviewLimit = 10
	}
	g, err := r.dao.GetChatGroupProfile(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ChatGroupProfileItem{}, nil, ErrGroupNotExist
		}
		r.log.Error(err)
		return ChatGroupProfileItem{}, nil, ErrDefault
	}
	members, err := r.dao.ListChatGroupMembersPreviewWithProfile(ctx, dao.ListChatGroupMembersPreviewWithProfileParams{
		GroupID: gid,
		Limit:   int32(memberPreviewLimit),
	})
	if err != nil {
		r.log.Error(err)
		return ChatGroupProfileItem{}, nil, ErrDefault
	}
	memberItems := make([]ChatGroupMemberProfile, 0, len(members))
	for _, v := range members {
		memberItems = append(memberItems, ChatGroupMemberProfile{
			UserID:   v.UserID,
			Username: v.Username,
			Avatar:   v.Avatar,
			Role:     v.Role,
			JoinTime: v.JoinTime,
		})
	}
	return ChatGroupProfileItem{
		GroupID:     g.GroupID,
		Name:        g.Name,
		Avatar:      g.Avatar,
		Description: g.Description,
		MemberCount: g.MemberCount,
		MemberLimit: g.MemberLimit,
		Ctime:       g.Ctime,
		Utime:       g.Utime,
		OwnerUserID: g.OwnerUserID,
		OwnerName:   g.OwnerUsername,
		OwnerAvatar: g.OwnerAvatar,
	}, memberItems, nil
}

func (r *ChatRepo) ListMembersWithProfile(ctx context.Context, groupID string, page, pageSize int) ([]ChatGroupMemberProfile, bool, error) {
	var gid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return nil, false, ErrParamsType
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListChatGroupMembersWithProfile(ctx, dao.ListChatGroupMembersWithProfileParams{
		GroupID: gid,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	hasMore := len(rows) >= int(limit)
	items := make([]ChatGroupMemberProfile, 0, pageSize)
	for i, v := range rows {
		if i >= pageSize {
			break
		}
		items = append(items, ChatGroupMemberProfile{
			UserID:   v.UserID,
			Username: v.Username,
			Avatar:   v.Avatar,
			Role:     v.Role,
			JoinTime: v.JoinTime,
		})
	}
	return items, hasMore, nil
}
