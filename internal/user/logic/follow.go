package logic

import (
	"context"
	"errors"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
)

func (ul *UserLogic) Follow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error) {
	var resp dto.FollowResp
	target := uri.TargetUserID
	if target == "" || target == userID {
		return resp, ErrParamsType
	}
	_, err := ul.userRepo.GetUserByID(ctx, target)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if err := ul.userRepo.FollowUser(ctx, userID, target); err != nil {
		if errors.Is(err, repo.ErrDefault) {
			ul.log.Error(err)
			return resp, ErrDefault
		}
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) Unfollow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error) {
	var resp dto.FollowResp
	target := uri.TargetUserID
	if target == "" || target == userID {
		return resp, ErrParamsType
	}
	if err := ul.userRepo.UnfollowUser(ctx, userID, target); err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) ListFollowing(ctx context.Context, userID string, req dto.FollowingListReq) (dto.FollowingListResp, error) {
	var resp dto.FollowingListResp
	viewID := userID
	if req.UserID != "" {
		viewID = req.UserID
	}
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.ListFollowing(ctx, viewID, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.FollowingUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.FollowingUserItem{
			UserID:     r.UserID,
			Username:   r.Username,
			Avatar:     r.Avatar,
			FollowTime: r.FollowTime,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}

func (ul *UserLogic) ListFollowers(ctx context.Context, userID string, req dto.FollowersListReq) (dto.FollowersListResp, error) {
	var resp dto.FollowersListResp
	viewID := userID
	if req.UserID != "" {
		viewID = req.UserID
	}
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.ListFollowers(ctx, viewID, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.FollowingUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.FollowingUserItem{
			UserID:     r.UserID,
			Username:   r.Username,
			Avatar:     r.Avatar,
			FollowTime: r.FollowTime,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}
