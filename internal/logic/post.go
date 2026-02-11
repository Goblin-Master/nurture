package logic

import (
	"context"
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
	List(ctx context.Context, req dto.PostListReq) (dto.PostListResp, error)
	Detail(ctx context.Context, req dto.PostDetailReq) (dto.PostDetailResp, error)
	NewPost(ctx context.Context, userID string, req dto.CreatePostReq) (dto.CreatePostResp, error)
	Publish(ctx context.Context, userID string, req dto.PublishPostReq) (dto.PublishPostResp, error)
}

type PostLogic struct {
	postRepo repo.IPostRepo
}

func NewPostLogic() *PostLogic {
	return &PostLogic{
		postRepo: repo.NewPostRepo(),
	}
}

var _ IPostLogic = (*PostLogic)(nil)

func (l *PostLogic) List(ctx context.Context, req dto.PostListReq) (dto.PostListResp, error) {
	var resp dto.PostListResp
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	var excludes []string
	if req.ExcludeIDs != "" {
		for _, v := range strings.Split(req.ExcludeIDs, ",") {
			s := strings.TrimSpace(v)
			if s != "" {
				excludes = append(excludes, s)
			}
		}
	}
	items, hasMore, err := l.postRepo.List(ctx, repo.PostListFilter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Status:     req.Status,
		TagID:      req.TagID,
		AuthorID:   req.AuthorID,
		OrderBy:    req.OrderBy,
		Order:      req.Order,
		Keyword:    req.Keyword,
		Strategy:   req.Strategy,
		ExcludeIDs: excludes,
	})
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.Items = make([]dto.PostItem, 0, len(items))
	for _, v := range items {
		y, m, ageText := calcAge(v.Birthday, time.Now())
		preview := makePreview(v.Content, 15)
		resp.Items = append(resp.Items, dto.PostItem{
			PostID:         v.PostID,
			AuthorID:       v.AuthorID,
			AuthorName:     v.AuthorName,
			AuthorAvatar:   v.AuthorAvatar,
			AuthorProvince: v.AuthorProvince,
			AuthorCity:     v.AuthorCity,
			Title:          v.Title,
			Content:        v.Content,
			ContentPreview: preview,
			Status:         v.Status,
			LikeCount:      v.LikeCount,
			DislikeCount:   v.DislikeCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			Cover:          v.Cover,
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
func (l *PostLogic) NewPost(ctx context.Context, userID string, req dto.CreatePostReq) (dto.CreatePostResp, error) {
	var resp dto.CreatePostResp
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		return resp, ErrParamsType
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	postID := uuid.NewString()
	now := time.Now().UnixMilli()
	err := l.postRepo.CreatePost(ctx, postID, userID, title, content, status, req.Cover, now, now, req.TagIDs)
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
		Content:        row.Content,
		Status:         row.Status,
		LikeCount:      row.LikeCount,
		DislikeCount:   row.DislikeCount,
		CollectCount:   row.CollectCount,
		CommentCount:   row.CommentCount,
		Cover:          row.Cover,
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
