package logic

import (
	"context"
	"nurture/internal/chat/dto"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (l *ChatLogic) CreateGroup(ctx context.Context, userID string, req dto.CreateChatGroupReq) (dto.CreateChatGroupResp, error) {
	var resp dto.CreateChatGroupResp
	if strings.TrimSpace(req.Name) == "" {
		return resp, ErrParamsType
	}
	if req.MemberLimit <= 0 || req.MemberLimit > 1000 {
		return resp, ErrParamsType
	}
	now := time.Now().UnixMilli()
	groupID := uuid.NewString()
	if err := l.chatRepo.CreateGroup(ctx, groupID, userID, req.Name, req.Avatar, req.Description, req.MemberLimit, now); err != nil {
		return resp, mapRepoErr(err)
	}
	resp.GroupID = groupID
	resp.Message = "OK"
	return resp, nil
}

func (l *ChatLogic) JoinGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error {
	if strings.TrimSpace(uri.GroupID) == "" {
		return ErrParamsType
	}
	now := time.Now().UnixMilli()
	if err := l.chatRepo.JoinGroup(ctx, uri.GroupID, userID, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) LeaveGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error {
	if strings.TrimSpace(uri.GroupID) == "" {
		return ErrParamsType
	}
	now := time.Now().UnixMilli()
	if err := l.chatRepo.LeaveGroup(ctx, uri.GroupID, userID, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) TransferOwner(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupTransferReq) error {
	if strings.TrimSpace(uri.GroupID) == "" || strings.TrimSpace(req.TargetUserID) == "" {
		return ErrParamsType
	}
	if req.TargetUserID == userID {
		return ErrParamsType
	}
	now := time.Now().UnixMilli()
	if err := l.chatRepo.TransferOwner(ctx, uri.GroupID, userID, req.TargetUserID, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) DissolveGroup(ctx context.Context, userID string, uri dto.ChatGroupIDUri) error {
	if strings.TrimSpace(uri.GroupID) == "" {
		return ErrParamsType
	}
	now := time.Now().UnixMilli()
	if err := l.chatRepo.DissolveGroup(ctx, uri.GroupID, userID, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (l *ChatLogic) ListMyGroups(ctx context.Context, userID string) (dto.ListMyChatGroupsResp, error) {
	var resp dto.ListMyChatGroupsResp
	rows, err := l.chatRepo.ListMyGroups(ctx, userID)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Items = make([]dto.ChatGroupListItem, 0, len(rows))
	for _, v := range rows {
		resp.Items = append(resp.Items, dto.ChatGroupListItem{
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
			LastMessageFromUserID: v.LastMessageFromUserID,
			LastMessageFromName:   v.LastMessageFromName,
			LastMessageType:       v.LastMessageType,
			LastMessageContent:    v.LastMessageContent,
			LastMessageTime:       v.LastMessageTime,
		})
	}
	return resp, nil
}

func (l *ChatLogic) DiscoverGroups(ctx context.Context, userID string, req dto.ChatGroupDiscoverReq) (dto.ChatGroupDiscoverResp, error) {
	var resp dto.ChatGroupDiscoverResp
	seed := strings.TrimSpace(req.Seed)
	if seed == "" {
		seed = uuid.NewString()
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cursorSortKey := ""
	cursorGroupID := ""
	if strings.TrimSpace(req.Cursor) != "" {
		sk, gid, err := parseDiscoverCursor(req.Cursor)
		if err != nil {
			return resp, err
		}
		cursorSortKey = sk
		cursorGroupID = gid
	}
	items, nextCursor, hasMore, err := l.chatRepo.ListDiscoverGroups(ctx, userID, seed, cursorSortKey, cursorGroupID, limit)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Seed = seed
	resp.HasMore = hasMore
	resp.NextCursor = nextCursor
	resp.Items = make([]dto.ChatGroupDiscoverItem, 0, len(items))
	for _, v := range items {
		resp.Items = append(resp.Items, dto.ChatGroupDiscoverItem{
			GroupID:     v.GroupID,
			Name:        v.Name,
			Avatar:      v.Avatar,
			MemberCount: v.MemberCount,
			MemberLimit: v.MemberLimit,
		})
	}
	return resp, nil
}

func (l *ChatLogic) SearchGroups(ctx context.Context, userID string, req dto.ChatGroupSearchReq) (dto.ChatGroupSearchResp, error) {
	var resp dto.ChatGroupSearchResp
	items, err := l.chatRepo.SearchGroupsByName(ctx, req.Keyword, req.Limit)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Items = make([]dto.ChatGroupDiscoverItem, 0, len(items))
	for _, v := range items {
		resp.Items = append(resp.Items, dto.ChatGroupDiscoverItem{
			GroupID:     v.GroupID,
			Name:        v.Name,
			Avatar:      v.Avatar,
			MemberCount: v.MemberCount,
			MemberLimit: v.MemberLimit,
		})
	}
	return resp, nil
}

func (l *ChatLogic) GetGroupProfile(ctx context.Context, userID string, uri dto.ChatGroupIDUri) (dto.ChatGroupProfileResp, error) {
	var resp dto.ChatGroupProfileResp
	if strings.TrimSpace(uri.GroupID) == "" {
		return resp, ErrParamsType
	}
	g, members, err := l.chatRepo.GetGroupProfile(ctx, uri.GroupID, 10)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Group = dto.ChatGroupProfile{
		GroupID:     g.GroupID,
		Name:        g.Name,
		Avatar:      g.Avatar,
		Description: g.Description,
		MemberCount: g.MemberCount,
		MemberLimit: g.MemberLimit,
		Ctime:       g.Ctime,
		Utime:       g.Utime,
	}
	resp.Owner = dto.ChatGroupProfileOwner{
		UserID:   g.OwnerUserID,
		Username: g.OwnerName,
		Avatar:   g.OwnerAvatar,
	}
	resp.MembersPreview = make([]dto.ChatGroupMemberItem, 0, len(members))
	for _, v := range members {
		resp.MembersPreview = append(resp.MembersPreview, dto.ChatGroupMemberItem{
			UserID:   v.UserID,
			Username: v.Username,
			Avatar:   v.Avatar,
			Role:     v.Role,
			JoinTime: v.JoinTime,
		})
	}
	return resp, nil
}

func (l *ChatLogic) ListMembers(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupMemberListReq) (dto.ChatGroupMemberListResp, error) {
	var resp dto.ChatGroupMemberListResp
	if strings.TrimSpace(uri.GroupID) == "" {
		return resp, ErrParamsType
	}
	if _, err := l.chatRepo.GetMemberRole(ctx, uri.GroupID, userID); err != nil {
		return resp, mapRepoErr(err)
	}
	items, hasMore, err := l.chatRepo.ListMembersWithProfile(ctx, uri.GroupID, req.Page, req.PageSize)
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	if resp.Page <= 0 {
		resp.Page = 1
	}
	if resp.PageSize <= 0 || resp.PageSize > 100 {
		resp.PageSize = 20
	}
	resp.HasMore = hasMore
	resp.Items = make([]dto.ChatGroupMemberItem, 0, len(items))
	for _, v := range items {
		resp.Items = append(resp.Items, dto.ChatGroupMemberItem{
			UserID:   v.UserID,
			Username: v.Username,
			Avatar:   v.Avatar,
			Role:     v.Role,
			JoinTime: v.JoinTime,
		})
	}
	return resp, nil
}

func (l *ChatLogic) ListMessages(ctx context.Context, userID string, uri dto.ChatGroupIDUri, req dto.ChatGroupMessageListReq) (dto.ChatGroupMessageListResp, error) {
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
