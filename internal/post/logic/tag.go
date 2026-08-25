package logic

import (
	"context"
	"errors"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

func (l *PostLogic) AdminCreateTag(ctx context.Context, req dto.AdminTagCreateReq) (dto.AdminTagCreateResp, error) {
	var resp dto.AdminTagCreateResp
	name := strings.TrimSpace(req.Name)
	if name == "" || utf8.RuneCountInString(name) > 32 {
		return resp, ErrParamsType
	}
	desc := strings.TrimSpace(req.Description)
	tagID := uuid.NewString()
	now := time.Now().UnixMilli()
	row, err := l.postRepo.CreateTag(ctx, tagID, name, desc, now)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.TagID = row.TagID
	resp.Name = row.Name
	resp.Description = row.Description
	return resp, nil
}

func (l *PostLogic) AdminDeleteTag(ctx context.Context, uri dto.AdminTagDeleteUri) error {
	if strings.TrimSpace(uri.TagID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.DeleteTag(ctx, uri.TagID); err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return ErrParamsType
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) ListTags(ctx context.Context, req dto.TagListReq) (dto.TagListResp, error) {
	var resp dto.TagListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListTags(ctx, strings.TrimSpace(req.Keyword), req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.TagItem, 0, len(items))
	for _, v := range items {
		resp.Items = append(resp.Items, dto.TagItem{
			TagID:       v.TagID,
			Name:        v.Name,
			Description: v.Description,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}
