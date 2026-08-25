package repo

import (
	"context"
	"errors"
	"fmt"
	postconstant "nurture/internal/post/constant"
	"nurture/internal/post/repo/cache"
	"nurture/internal/post/repo/dao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tmc/langchaingo/schema"
)

func (r *PostRepo) listRecommend(ctx context.Context, userID string, page, pageSize int) ([]PostRow, bool, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, false, ErrParamsType
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, ErrParamsType
	}
	profileText, err := r.dao.GetUserRecommendProfile(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrParamsType
		}
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	profileText = strings.TrimSpace(profileText)
	if profileText == "" {
		return nil, false, ErrParamsType
	}
	if r.ai == nil || !r.ai.EmbeddingEnabled() {
		return nil, false, ErrParamsType
	}
	topK := 200
	docs, err := r.ai.SimilaritySearch(ctx, profileText, []string{postconstant.RecommendCollection}, topK)
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	postIDs := extractPostIDs(docs)
	if len(postIDs) == 0 {
		return nil, false, ErrParamsType
	}
	start := (page - 1) * pageSize
	if start >= len(postIDs) {
		return []PostRow{}, false, nil
	}
	end := start + pageSize
	if end > len(postIDs) {
		end = len(postIDs)
	}
	idsPage := postIDs[start:end]
	arr := make([]pgtype.UUID, 0, len(idsPage))
	for _, id := range idsPage {
		var pid pgtype.UUID
		if err := pid.Scan(id); err != nil {
			continue
		}
		arr = append(arr, pid)
	}
	if len(arr) == 0 {
		return []PostRow{}, end < len(postIDs), nil
	}
	rows, err := r.dao.ListPostsByIDs(ctx, dao.ListPostsByIDsParams{
		Column1: arr,
		Column2: userID,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]PostRow, 0, len(rows))
	for _, v := range rows {
		res = append(res, toPostRow(
			v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
			v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	return res, end < len(postIDs), nil
}

func extractPostIDs(docs []schema.Document) []string {
	seen := make(map[string]struct{}, len(docs))
	out := make([]string, 0, len(docs))
	for _, d := range docs {
		v, ok := d.Metadata["post_id"]
		if !ok {
			continue
		}
		var id string
		switch t := v.(type) {
		case string:
			id = t
		case []byte:
			id = string(t)
		default:
			continue
		}
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (r *PostRepo) TouchUserRecommendProfile(ctx context.Context, userID string, postID string) error {
	var uid, pid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return ErrParamsType
	}
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	row, err := r.dao.GetPostRecommendDoc(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPostNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	var tags string
	switch v := row.Tags.(type) {
	case string:
		tags = v
	case []byte:
		tags = string(v)
	default:
		tags = ""
	}
	snippet := strings.TrimSpace(fmt.Sprintf("%s %s", row.Title, tags))
	if snippet == "" {
		return nil
	}
	now := time.Now().UnixMilli()
	err = r.dao.UpsertUserRecommendProfile(ctx, dao.UpsertUserRecommendProfileParams{
		UserID:      uid,
		ProfileText: snippet,
		Utime:       now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *PostRepo) IndexPostForRecommend(ctx context.Context, postID string) error {
	if r.ai == nil || !r.ai.EmbeddingEnabled() {
		return nil
	}
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	row, err := r.dao.GetPostRecommendDoc(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPostNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	var tags string
	switch v := row.Tags.(type) {
	case string:
		tags = v
	case []byte:
		tags = string(v)
	default:
		tags = ""
	}
	doc := strings.TrimSpace(fmt.Sprintf("%s\n%s\n%s", row.Title, tags, row.Content))
	if doc == "" {
		return nil
	}
	err = r.ai.AddDocumentWithMetadata(ctx, postconstant.RecommendCollection, doc, map[string]any{
		"post_id": row.PostID,
	}, false)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *PostRepo) TouchUserTagPref(ctx context.Context, userID string, postID string, score float64) error {
	if r.rdb == nil {
		return nil
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(postID) == "" {
		return ErrParamsType
	}
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	tags, err := r.dao.ListTagNamesByPost(ctx, pid)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if len(tags) == 0 {
		return nil
	}
	key := cache.UserTagPrefKey(userID)
	pipe := r.rdb.TxPipeline()
	for _, t := range tags {
		name := strings.TrimSpace(t)
		if name == "" {
			continue
		}
		pipe.ZIncrBy(ctx, key, score, name)
	}
	pipe.Expire(ctx, key, time.Duration(postconstant.UserTagPrefTTL)*time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
