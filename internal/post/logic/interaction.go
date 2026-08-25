package logic

import (
	"context"
	"errors"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"
	"strings"

	"github.com/google/uuid"
)

func (l *PostLogic) LikePost(ctx context.Context, userID string, uri dto.PostDetailReq) error {
	if strings.TrimSpace(uri.PostID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.LikePost(ctx, uri.PostID, userID); err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	if err := l.postRepo.TouchUserRecommendProfile(ctx, userID, uri.PostID); err != nil {
		l.log.Error(err)
	}
	if err := l.postRepo.TouchUserTagPref(ctx, userID, uri.PostID, 3); err != nil {
		l.log.Error(err)
	}
	return nil
}

func (l *PostLogic) UnlikePost(ctx context.Context, userID string, uri dto.PostDetailReq) error {
	if strings.TrimSpace(uri.PostID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.UnlikePost(ctx, uri.PostID, userID); err != nil {
		l.log.Error(err)
		return ErrDefault
	}
	if err := l.postRepo.TouchUserTagPref(ctx, userID, uri.PostID, -3); err != nil {
		l.log.Error(err)
	}
	return nil
}

func (l *PostLogic) CollectPost(ctx context.Context, userID string, uri dto.PostDetailReq) (dto.CollectResp, error) {
	var resp dto.CollectResp
	if strings.TrimSpace(uri.PostID) == "" {
		return resp, ErrParamsType
	}
	cid := uuid.NewString()
	if err := l.postRepo.CollectPost(ctx, uri.PostID, userID, cid); err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return resp, ErrInvalidPostStatus
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	if err := l.postRepo.TouchUserRecommendProfile(ctx, userID, uri.PostID); err != nil {
		l.log.Error(err)
	}
	if err := l.postRepo.TouchUserTagPref(ctx, userID, uri.PostID, 4); err != nil {
		l.log.Error(err)
	}
	resp.CollectionID = cid
	resp.Message = "OK"
	return resp, nil
}

func (l *PostLogic) UncollectPost(ctx context.Context, userID string, uri dto.PostDetailReq) (dto.CollectResp, error) {
	var resp dto.CollectResp
	if strings.TrimSpace(uri.PostID) == "" {
		return resp, ErrParamsType
	}
	if err := l.postRepo.UncollectPost(ctx, uri.PostID, userID); err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	if err := l.postRepo.TouchUserTagPref(ctx, userID, uri.PostID, -4); err != nil {
		l.log.Error(err)
	}
	resp.Message = "OK"
	return resp, nil
}
