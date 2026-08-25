package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

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
		l.log.Error(err)
		return resp, ErrDefault
	}
	if err := l.postRepo.IndexPostForRecommend(ctx, req.PostID); err != nil {
		l.log.Error(err)
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
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.PostID = uri.PostID
	resp.Status = "draft"
	resp.Message = "更新成功"
	return resp, nil
}

func (l *PostLogic) DeleteDraft(ctx context.Context, userID string, uri dto.PostDetailReq) error {
	if strings.TrimSpace(uri.PostID) == "" {
		return ErrParamsType
	}
	err := l.postRepo.DeleteDraft(ctx, uri.PostID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return ErrPostNotExist
		}
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (l *PostLogic) DeletePost(ctx context.Context, userID string, uri dto.PostDetailReq) error {
	if strings.TrimSpace(uri.PostID) == "" {
		return ErrParamsType
	}
	err := l.postRepo.DeletePost(ctx, uri.PostID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return ErrPostNotExist
		}
		if errors.Is(err, repo.ErrInvalidPostStatus) {
			return ErrInvalidPostStatus
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
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
		l.log.Error(err)
		return resp, ErrDefault
	}
	if status == "published" || status == "milestone" {
		if err := l.postRepo.IndexPostForRecommend(ctx, postID); err != nil {
			l.log.Error(err)
		}
	}
	resp.PostID = postID
	resp.Status = status
	resp.Message = "创建成功"
	return resp, nil
}

func (l *PostLogic) Detail(ctx context.Context, userID string, req dto.PostDetailReq) (dto.PostDetailResp, error) {
	var resp dto.PostDetailResp
	if strings.TrimSpace(req.PostID) == "" {
		return resp, ErrParamsType
	}
	row, err := l.postRepo.GetDetail(ctx, userID, req.PostID)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotExist) {
			return resp, ErrPostNotExist
		}
		l.log.Error(err)
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
		IsLike:         row.IsLike,
		IsDislike:      row.IsDislike,
		IsCollect:      row.IsCollect,
		Tags:           row.Tags,
		BabyAgeYear:    y,
		BabyAgeMonth:   m,
		BabyAgeText:    ageText,
	}
	if userID != "" && userID != row.AuthorID && l.followReader != nil {
		ok, e := l.followReader.IsFollowing(ctx, userID, row.AuthorID)
		if e != nil {
			l.log.Error(e)
		} else {
			resp.Post.IsFollow = ok
		}
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
