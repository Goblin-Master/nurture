package logic

import (
	"context"
	"errors"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (l *PostLogic) DeleteComment(ctx context.Context, userID string, uri dto.CommentDeleteReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.DeleteComment(ctx, uri.CommentID, userID); err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) LikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.LikeComment(ctx, uri.CommentID, userID); err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) UnlikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.UnlikeComment(ctx, uri.CommentID, userID); err != nil {
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) UpdateComment(ctx context.Context, userID string, uri dto.CommentUpdateReq, req dto.UpdateCommentReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if len(req.Content) == 0 {
		return ErrParamsType
	}
	err := l.postRepo.UpdateComment(ctx, uri.CommentID, userID, string(req.Content))
	if err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) ListReplies(ctx context.Context, userID string, uri dto.CommentRepliesReq, req dto.CommentListReq) (dto.CommentListResp, error) {
	var resp dto.CommentListResp
	if strings.TrimSpace(uri.PostID) == "" || strings.TrimSpace(uri.CommentID) == "" {
		return resp, ErrParamsType
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	status, err := l.postRepo.GetPostStatus(ctx, uri.PostID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	pPostID, pStatus, err := l.postRepo.GetCommentParentInfo(ctx, uri.CommentID)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	if pPostID != uri.PostID {
		return resp, ErrParamsType
	}
	if pStatus == "hidden" {
		return resp, ErrInvalidPostStatus
	}
	rows, hasMore, err := l.postRepo.ListRepliesByComment(ctx, uri.CommentID, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.CommentItem, 0, len(rows))
	for _, v := range rows {
		resp.Items = append(resp.Items, dto.CommentItem{
			CommentID:  v.CommentID,
			UserID:     v.UserID,
			Username:   v.Username,
			Avatar:     v.Avatar,
			Content:    v.Content,
			LikeCount:  v.LikeCount,
			ReplyCount: v.ReplyCount,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
			HasLiked:   v.HasLiked,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}

func (l *PostLogic) CreateComment(ctx context.Context, userID string, postID string, req dto.CreateCommentReq) (dto.CreateCommentResp, error) {
	var resp dto.CreateCommentResp
	if strings.TrimSpace(postID) == "" || len(req.Content) == 0 {
		return resp, ErrParamsType
	}
	status, err := l.postRepo.GetPostStatus(ctx, postID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	parentID := strings.TrimSpace(req.ParentID)
	if parentID != "" {
		pPostID, pStatus, err := l.postRepo.GetCommentParentInfo(ctx, parentID)
		if err != nil {
			l.log.Error(err)
			return resp, ErrDefault
		}
		if pPostID != postID {
			return resp, ErrParamsType
		}
		if pStatus != "visible" {
			return resp, ErrInvalidPostStatus
		}
	}
	commentID := uuid.NewString()
	now := time.Now().UnixMilli()
	if err := l.postRepo.CreateComment(ctx, commentID, postID, userID, ifNonEmptyPtr(parentID), string(req.Content), now); err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	if err := l.postRepo.TouchUserRecommendProfile(ctx, userID, postID); err != nil {
		l.log.Error(err)
	}
	if err := l.postRepo.TouchUserTagPref(ctx, userID, postID, 5); err != nil {
		l.log.Error(err)
	}
	resp.CommentID = commentID
	resp.Message = "创建成功"
	return resp, nil
}

func ifNonEmptyPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func (l *PostLogic) ListComments(ctx context.Context, userID string, postID string, req dto.CommentListReq) (dto.CommentListResp, error) {
	var resp dto.CommentListResp
	if strings.TrimSpace(postID) == "" {
		return resp, ErrParamsType
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	status, err := l.postRepo.GetPostStatus(ctx, postID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	rows, hasMore, err := l.postRepo.ListCommentsByPost(ctx, postID, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.CommentItem, 0, len(rows))
	for _, v := range rows {
		resp.Items = append(resp.Items, dto.CommentItem{
			CommentID:  v.CommentID,
			UserID:     v.UserID,
			Username:   v.Username,
			Avatar:     v.Avatar,
			Content:    v.Content,
			LikeCount:  v.LikeCount,
			ReplyCount: v.ReplyCount,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
			HasLiked:   v.HasLiked,
		})
	}
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}
