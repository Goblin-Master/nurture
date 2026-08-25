package logic

import (
	"context"
	"errors"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
)

// admin
func (ul *UserLogic) AdminListUsers(ctx context.Context, req dto.AdminListUsersReq) (dto.AdminListUsersResp, error) {
	var resp dto.AdminListUsersResp
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.AdminListUsers(ctx, req.Keyword, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.AdminUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.AdminUserItem{
			UserID:   r.UserID,
			Username: r.Username,
			Avatar:   r.Avatar,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}

// admin
func (ul *UserLogic) AdminPromoteToAdmin(ctx context.Context, userID string) (string, error) {
	err := ul.userRepo.AdminUpdateUserRole(ctx, userID, int16(jwtx.ADMIN))
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return "", ErrUserNotExist
		}
		ul.log.Error(err)
		return "", ErrDefault
	}
	return "OK", nil
}
