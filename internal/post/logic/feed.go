package logic

import (
	"context"
	"encoding/json"
	"errors"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"
	"time"
)

func (l *PostLogic) Home(ctx context.Context, userID string, req dto.PostHomeListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListHome(ctx, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) Following(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListFollowing(ctx, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) ListMyCollections(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListMyCollections(ctx, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) ListByTag(ctx context.Context, userID string, req dto.PostTagListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListByTag(ctx, userID, req.TagID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) Search(ctx context.Context, userID string, req dto.PostSearchListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.Search(ctx, userID, req.Keyword, req.TagID, req.Strategy, req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) ListMyPosts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListByAuthor(ctx, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) ListMyDrafts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListDraftsByAuthor(ctx, userID, req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			IsLike:         v.IsLike,
			IsDislike:      v.IsDislike,
			IsCollect:      v.IsCollect,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) ListMyMilestones(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListMilestonesByAuthor(ctx, userID, req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        json.RawMessage([]byte(v.Content)),
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Ctime:          v.Ctime,
			Utime:          v.Utime,
			Tags:           v.Tags,
			BabyAgeYear:    y,
			BabyAgeMonth:   m,
			BabyAgeText:    ageText,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}
