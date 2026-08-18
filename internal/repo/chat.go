package repo

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/chat/repo/dao"
	"nurture/internal/global"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChatGroupListItem struct {
	GroupID               string
	Name                  string
	Avatar                string
	Description           string
	MemberLimit           int32
	MemberCount           int32
	Role                  string
	UnreadCount           int64
	Ctime                 int64
	Utime                 int64
	LastMessageFromUserID string
	LastMessageFromName   string
	LastMessageType       string
	LastMessageContent    string
	LastMessageTime       int64
}

type ChatGroupMemberProfile struct {
	UserID   string
	Username string
	Avatar   string
	Role     string
	JoinTime int64
}

type ChatGroupMessageItem struct {
	MessageID  string
	GroupID    string
	FromUserID string
	Type       string
	Content    string
	Ctime      int64
}

type ChatGroupDiscoverItem struct {
	GroupID     string
	Name        string
	Avatar      string
	MemberCount int32
	MemberLimit int32
	SortKey     string
}

type ChatGroupProfileItem struct {
	GroupID     string
	Name        string
	Avatar      string
	Description string
	MemberCount int32
	MemberLimit int32
	Ctime       int64
	Utime       int64
	OwnerUserID string
	OwnerName   string
	OwnerAvatar string
}

type IChatRepo interface {
	CreateGroup(ctx context.Context, groupID, ownerID, name, avatar, description string, memberLimit int32, now int64) error
	JoinGroup(ctx context.Context, groupID, userID string, now int64) error
	LeaveGroup(ctx context.Context, groupID, userID string, now int64) error
	TransferOwner(ctx context.Context, groupID, ownerID, targetUserID string, now int64) error
	DissolveGroup(ctx context.Context, groupID, ownerID string, now int64) error
	UpdateMemberLastSeenTime(ctx context.Context, groupID, userID string, lastSeenTime int64) error

	ListMyGroups(ctx context.Context, userID string) ([]ChatGroupListItem, error)
	ListDiscoverGroups(ctx context.Context, userID, seed, cursorSortKey, cursorGroupID string, limit int) ([]ChatGroupDiscoverItem, string, bool, error)
	SearchGroupsByName(ctx context.Context, keyword string, limit int) ([]ChatGroupDiscoverItem, error)
	GetGroupProfile(ctx context.Context, groupID string, memberPreviewLimit int) (ChatGroupProfileItem, []ChatGroupMemberProfile, error)
	ListMembersWithProfile(ctx context.Context, groupID string, page, pageSize int) ([]ChatGroupMemberProfile, bool, error)

	SaveMessage(ctx context.Context, groupID, messageID, fromUserID, msgType, content string, now int64) error
	ListMessagesLatest(ctx context.Context, groupID string, limit int) ([]ChatGroupMessageItem, error)
	ListMessagesBefore(ctx context.Context, groupID string, beforeCtime int64, beforeMessageID string, limit int) ([]ChatGroupMessageItem, error)
	ListMessagesAfter(ctx context.Context, groupID string, afterCtime int64, afterMessageID string, limit int) ([]ChatGroupMessageItem, error)

	GetMemberRole(ctx context.Context, groupID, userID string) (string, error)
}

type ChatRepo struct {
	dao *dao.Queries
}

func NewChatRepo() *ChatRepo {
	return &ChatRepo{
		dao: dao.New(global.DB),
	}
}

var _ IChatRepo = (*ChatRepo)(nil)

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
		global.Log.Error(err)
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

	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatGroupNotExist
		}
		global.Log.Error(err)
		return ErrDefault
	}

	if _, e := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: uid}); e == nil {
		return tx.Commit(ctx)
	} else if !errors.Is(e, pgx.ErrNoRows) {
		global.Log.Error(e)
		return ErrDefault
	}

	if g.MemberCount >= g.MemberLimit {
		return ErrChatGroupFull
	}

	if err := qtx.CreateChatGroupMember(ctx, dao.CreateChatGroupMemberParams{
		GroupID: gid,
		UserID:  uid,
		Role:    "member",
		Ctime:   now,
	}); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if _, err := qtx.IncChatGroupMemberCount(ctx, dao.IncChatGroupMemberCountParams{
		GroupID: gid,
		Utime:   now,
	}); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	return tx.Commit(ctx)
}

func (r *ChatRepo) LeaveGroup(ctx context.Context, groupID, userID string, now int64) error {
	var gid, uid pgtype.UUID
	if err := gid.Scan(groupID); err != nil {
		return ErrParamsType
	}
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}

	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	if _, err := qtx.LockChatGroupByID(ctx, gid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatGroupNotExist
		}
		global.Log.Error(err)
		return ErrDefault
	}
	role, err := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatNotMember
		}
		global.Log.Error(err)
		return ErrDefault
	}
	if role == "owner" {
		return ErrChatOwnerMustTransferFirst
	}
	aff, err := qtx.DeleteChatGroupMember(ctx, dao.DeleteChatGroupMemberParams{
		GroupID: gid,
		UserID:  uid,
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrChatNotMember
	}
	if _, err := qtx.DecChatGroupMemberCount(ctx, dao.DecChatGroupMemberCountParams{
		GroupID: gid,
		Utime:   now,
	}); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	return tx.Commit(ctx)
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

	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatGroupNotExist
		}
		global.Log.Error(err)
		return ErrDefault
	}
	if g.OwnerID != oid {
		return ErrChatPermissionDenied
	}
	if _, err := qtx.GetChatGroupMemberRole(ctx, dao.GetChatGroupMemberRoleParams{GroupID: gid, UserID: tid}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatNotMember
		}
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
		return ErrDefault
	}
	return tx.Commit(ctx)
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
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	qtx := r.dao.WithTx(tx)
	g, err := qtx.LockChatGroupByID(ctx, gid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrChatGroupNotExist
		}
		global.Log.Error(err)
		return ErrDefault
	}
	if g.OwnerID != oid {
		return ErrChatPermissionDenied
	}
	if _, err := qtx.DeleteChatGroupMessagesByGroupID(ctx, gid); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if _, err := qtx.DeleteChatGroupMembersByGroupID(ctx, gid); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	aff, err := qtx.DeleteChatGroupByID(ctx, gid)
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrChatGroupNotExist
	}
	return tx.Commit(ctx)
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
		global.Log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrChatNotMember
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
		global.Log.Error(err)
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
			global.Log.Error(err)
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
			global.Log.Error(err)
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
		global.Log.Error(err)
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
			return ChatGroupProfileItem{}, nil, ErrChatGroupNotExist
		}
		global.Log.Error(err)
		return ChatGroupProfileItem{}, nil, ErrDefault
	}
	members, err := r.dao.ListChatGroupMembersPreviewWithProfile(ctx, dao.ListChatGroupMembersPreviewWithProfileParams{
		GroupID: gid,
		Limit:   int32(memberPreviewLimit),
	})
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
			return "", ErrChatNotMember
		}
		global.Log.Error(err)
		return "", ErrDefault
	}
	return role, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}
