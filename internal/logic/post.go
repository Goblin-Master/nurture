package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/repo"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type IPostLogic interface {
	Home(ctx context.Context, req dto.PostHomeListReq) (dto.PostListResp, error)
	ListByTag(ctx context.Context, req dto.PostTagListReq) (dto.PostListResp, error)
	Search(ctx context.Context, req dto.PostSearchListReq) (dto.PostListResp, error)
	ListMyPosts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	ListMyDrafts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	ListMyMilestones(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	Following(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	Detail(ctx context.Context, req dto.PostDetailReq) (dto.PostDetailResp, error)
	NewPost(ctx context.Context, userID string, req dto.CreatePostReq) (dto.CreatePostResp, error)
	Publish(ctx context.Context, userID string, req dto.PublishPostReq) (dto.PublishPostResp, error)
	UpdateDraft(ctx context.Context, userID string, uri dto.PostDetailReq, req dto.UpdateDraftReq) (dto.UpdateDraftResp, error)
	CreateComment(ctx context.Context, userID string, postID string, req dto.CreateCommentReq) (dto.CreateCommentResp, error)
	ListComments(ctx context.Context, userID string, postID string, req dto.CommentListReq) (dto.CommentListResp, error)
	ListReplies(ctx context.Context, userID string, uri dto.CommentRepliesReq, req dto.CommentListReq) (dto.CommentListResp, error)
	DeleteComment(ctx context.Context, userID string, uri dto.CommentDeleteReq) error
	UpdateComment(ctx context.Context, userID string, uri dto.CommentUpdateReq, req dto.UpdateCommentReq) error
	LikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error
	UnlikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error
}

type PostLogic struct {
	postRepo *repo.PostRepo
}

func NewPostLogic() *PostLogic {
	return &PostLogic{
		postRepo: repo.NewPostRepo(),
	}
}

var _ IPostLogic = (*PostLogic)(nil)

func (l *PostLogic) Home(ctx context.Context, req dto.PostHomeListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListHome(ctx, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		global.Log.Error(err)
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
		global.Log.Error(err)
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

func (l *PostLogic) DeleteComment(ctx context.Context, userID string, uri dto.CommentDeleteReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.DeleteComment(ctx, uri.CommentID, userID); err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		global.Log.Error(err)
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
		global.Log.Error(err)
		return ErrDefault
	}
	return nil
}
func (l *PostLogic) UnlikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error {
	if strings.TrimSpace(uri.CommentID) == "" {
		return ErrParamsType
	}
	if err := l.postRepo.UnlikeComment(ctx, uri.CommentID, userID); err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	pPostID, pStatus, err := l.postRepo.GetCommentParentInfo(ctx, uri.CommentID)
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
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

func (l *PostLogic) ListByTag(ctx context.Context, req dto.PostTagListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.ListByTag(ctx, req.TagID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		global.Log.Error(err)
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

func (l *PostLogic) Search(ctx context.Context, req dto.PostSearchListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	items, hasMore, err := l.postRepo.Search(ctx, req.Keyword, req.TagID, req.Strategy, req.Page, req.PageSize)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return resp, ErrParamsType
		}
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
func (l *PostLogic) Publish(ctx context.Context, userID string, req dto.PublishPostReq) (dto.PublishPostResp, error) {
	var resp dto.PublishPostResp
	if strings.TrimSpace(req.PostID) == "" {
		return resp, ErrParamsType
	}
	err := l.postRepo.Publish(ctx, req.PostID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		if errors.Is(err, repo.ErrPostNotDraft) {
			return resp, ErrInvalidPostStatus
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.PostID = req.PostID
	resp.Status = "published"
	resp.Message = "发布成功"
	return resp, nil
}
func (l *PostLogic) UpdateDraft(ctx context.Context, userID string, uri dto.PostDetailReq, req dto.UpdateDraftReq) (dto.UpdateDraftResp, error) {
	var resp dto.UpdateDraftResp
	if strings.TrimSpace(uri.PostID) == "" {
		return resp, ErrParamsType
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len(req.Content) == 0 {
		return resp, ErrParamsType
	}
	err := l.postRepo.UpdateDraft(ctx, uri.PostID, userID, title, string(req.Content), req.TagIDs)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		if errors.Is(err, repo.ErrPostNotDraft) {
			return resp, ErrInvalidPostStatus
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.PostID = uri.PostID
	resp.Status = "draft"
	resp.Message = "更新成功"
	return resp, nil
}
func (l *PostLogic) NewPost(ctx context.Context, userID string, req dto.CreatePostReq) (dto.CreatePostResp, error) {
	var resp dto.CreatePostResp
	title := strings.TrimSpace(req.Title)
	if title == "" || len(req.Content) == 0 {
		return resp, ErrParamsType
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	postID := uuid.NewString()
	now := time.Now().UnixMilli()
	err := l.postRepo.CreatePost(ctx, postID, userID, title, string(req.Content), status, now, now, req.TagIDs)
	if err != nil {
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return resp, ErrInvalidPostStatus
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.PostID = postID
	resp.Status = status
	resp.Message = "创建成功"
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	parentID := strings.TrimSpace(req.ParentID)
	if parentID != "" {
		pPostID, pStatus, err := l.postRepo.GetCommentParentInfo(ctx, parentID)
		if err != nil {
			global.Log.Error(err)
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
		global.Log.Error(err)
		return resp, ErrDefault
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if status != "published" {
		return resp, ErrInvalidPostStatus
	}
	rows, hasMore, err := l.postRepo.ListCommentsByPost(ctx, postID, userID, req.Page, req.PageSize, req.Strategy)
	if err != nil {
		global.Log.Error(err)
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
func (l *PostLogic) Detail(ctx context.Context, req dto.PostDetailReq) (dto.PostDetailResp, error) {
	var resp dto.PostDetailResp
	if strings.TrimSpace(req.PostID) == "" {
		return resp, ErrParamsType
	}
	row, err := l.postRepo.GetDetail(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	y, m, ageText := calcAge(row.Birthday, time.Now())
	resp.Post = dto.PostDetail{
		PostID:         row.PostID,
		AuthorID:       row.AuthorID,
		AuthorName:     row.AuthorName,
		AuthorAvatar:   row.AuthorAvatar,
		AuthorProvince: row.AuthorProvince,
		AuthorCity:     row.AuthorCity,
		Title:          row.Title,
		Content:        json.RawMessage([]byte(row.Content)),
		Status:         row.Status,
		LikeCount:      row.LikeCount,
		DislikeCount:   row.DislikeCount,
		CollectCount:   row.CollectCount,
		CommentCount:   row.CommentCount,
		Ctime:          row.Ctime,
		Utime:          row.Utime,
		Tags:           row.Tags,
		BabyAgeYear:    y,
		BabyAgeMonth:   m,
		BabyAgeText:    ageText,
	}
	return resp, nil
}
func calcAge(birthdayMs int64, now time.Time) (int, int, string) {
	if birthdayMs <= 0 {
		return 0, 0, ""
	}
	b := time.UnixMilli(birthdayMs)
	y := now.Year() - b.Year()
	m := int(now.Month()) - int(b.Month())
	if now.Day() < b.Day() {
		m--
	}
	if m < 0 {
		y--
		m += 12
	}
	if y < 0 {
		y = 0
	}
	if m < 0 {
		m = 0
	}
	var text string
	if y == 0 {
		text = fmt.Sprintf("宝宝%d个月", m)
	} else {
		text = fmt.Sprintf("宝宝%d岁%d个月", y, m)
	}
	return y, m, text
}

func makePreview(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	rs := []rune(s)
	return string(rs[:n])
}
