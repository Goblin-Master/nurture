package repo

import (
	"context"
	"errors"
	postconstant "nurture/internal/post/constant"
	"nurture/internal/post/repo/cache"
	"nurture/internal/post/repo/dao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostRepo) GetDetail(ctx context.Context, userID, postID string) (PostRow, error) {
	key := cache.HotDetailKey(postID, userID)
	{
		var cached PostRow
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached, nil
		}
	}
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return PostRow{}, err
	}
	row, err := r.dao.GetPostDetail(ctx, dao.GetPostDetailParams{
		PostID:  pid,
		Column2: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostRow{}, ErrPostNotExist
		}
		r.log.Error(err)
		return PostRow{}, ErrDefault
	}
	var tags []string
	switch v := row.Tags.(type) {
	case []string:
		tags = v
	case []interface{}:
		for _, it := range v {
			if s, ok := it.(string); ok {
				tags = append(tags, s)
			}
		}
	default:
		tags = []string{}
	}
	var birthday int64
	switch v := row.Birthday.(type) {
	case int64:
		birthday = v
	case int32:
		birthday = int64(v)
	default:
		birthday = 0
	}
	ret := PostRow{
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
		Ctime:          row.Ctime,
		Utime:          row.Utime,
		Birthday:       birthday,
		Tags:           tags,
		IsLike:         row.IsLike,
		IsDislike:      row.IsDislike,
		IsCollect:      row.IsCollect,
	}
	_ = cache.SetJSON(ctx, r.rdb, key, ret, time.Duration(postconstant.HotDetailTTL)*time.Second)
	return ret, nil
}

func toPostRow(postID, authorID, authorName, authorAvatar, authorProvince, authorCity, title, content, status string,
	likeCount, dislikeCount, collectCount, commentCount int32,
	ctime, utime int64, birthday interface{}, tags interface{}, isLike, isDislike, isCollect bool) PostRow {
	var tagList []string
	switch v := tags.(type) {
	case []string:
		tagList = v
	case []interface{}:
		for _, it := range v {
			if s, ok := it.(string); ok {
				tagList = append(tagList, s)
			}
		}
	default:
		tagList = []string{}
	}
	var bday int64
	switch v := birthday.(type) {
	case int64:
		bday = v
	case int32:
		bday = int64(v)
	default:
		bday = 0
	}
	return PostRow{
		PostID:         postID,
		AuthorID:       authorID,
		AuthorName:     authorName,
		AuthorAvatar:   authorAvatar,
		AuthorProvince: authorProvince,
		AuthorCity:     authorCity,
		Title:          title,
		Content:        content,
		Status:         status,
		LikeCount:      likeCount,
		DislikeCount:   dislikeCount,
		CollectCount:   collectCount,
		CommentCount:   commentCount,
		Ctime:          ctime,
		Utime:          utime,
		Birthday:       bday,
		Tags:           tagList,
		IsLike:         isLike,
		IsDislike:      isDislike,
		IsCollect:      isCollect,
	}
}

func (r *PostRepo) CreatePost(ctx context.Context, postID, authorID, title, content, status string, ctime, utime int64, tagIDs []string) error {
	var pid, aid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := aid.Scan(authorID); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	err = qtx.CreatePost(ctx, dao.CreatePostParams{
		PostID:   pid,
		AuthorID: aid,
		Title:    title,
		Content:  content,
		Status:   status,
		Cover:    "",
		Ctime:    ctime,
		Utime:    utime,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" { // check_violation
			_ = tx.Rollback(ctx)
			return ErrInvalidPostStatus
		}
		if strings.Contains(err.Error(), "SQLSTATE 23514") {
			_ = tx.Rollback(ctx)
			return ErrInvalidPostStatus
		}
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	for _, tid := range tagIDs {
		if strings.TrimSpace(tid) == "" {
			continue
		}
		var tg pgtype.UUID
		if err := tg.Scan(tid); err != nil {
			continue
		}
		exists, err := qtx.TagExists(ctx, tg)
		if err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		if !exists {
			_ = tx.Rollback(ctx)
			return ErrParamsType
		}
		if err := qtx.AddPostTag(ctx, dao.AddPostTagParams{
			PostID: pid,
			TagID:  tg,
		}); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) DeleteDraft(ctx context.Context, postID, authorID string) error {
	var pid, aid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	if err := aid.Scan(authorID); err != nil {
		return ErrParamsType
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	aff, err := qtx.DeleteDraftByOwner(ctx, dao.DeleteDraftByOwnerParams{
		PostID:   pid,
		AuthorID: aid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff <= 0 {
		_ = tx.Rollback(ctx)
		return ErrPostNotExist
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) DeletePost(ctx context.Context, postID, authorID string) error {
	var pid, aid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	if err := aid.Scan(authorID); err != nil {
		return ErrParamsType
	}
	status, err := r.dao.GetPostStatusByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPostNotExist
		}
		r.log.Error(err)
		return ErrDefault
	}
	if status == "draft" {
		return ErrInvalidPostStatus
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	if err := qtx.DeleteCommentLikesByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCommentClosuresByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCommentsByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeletePostLikesByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCollectionsByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	aff, err := qtx.DeletePostByOwner(ctx, dao.DeletePostByOwnerParams{
		PostID:   pid,
		AuthorID: aid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff <= 0 {
		_ = tx.Rollback(ctx)
		return ErrPostNotExist
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = cache.ScanDel(ctx, r.rdb, cache.HotDetailPattern(postID), 200)
	_ = cache.ScanDel(ctx, r.rdb, cache.HotCommentsAllUsersPattern(postID), 200)
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListAllPattern(), 500)
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListByTagAllPattern(), 500)
	return nil
}

func (r *PostRepo) Publish(ctx context.Context, postID, userID string) error {
	var pid, uid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	count, err := r.dao.PublishPost(ctx, dao.PublishPostParams{
		PostID:   pid,
		AuthorID: uid,
		Utime:    time.Now().UnixMilli(),
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if count == 0 {
		return ErrPostNotDraft
	}
	return nil
}

func (r *PostRepo) UpdateDraft(ctx context.Context, postID, userID, title, content string, tagIDs []string) error {
	var pid, uid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	aff, err := qtx.UpdateDraftByOwner(ctx, dao.UpdateDraftByOwnerParams{
		PostID:   pid,
		AuthorID: uid,
		Title:    title,
		Content:  content,
		Utime:    time.Now().UnixMilli(),
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff == 0 {
		_ = tx.Rollback(ctx)
		return ErrPostNotDraft
	}
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	for _, tid := range tagIDs {
		if strings.TrimSpace(tid) == "" {
			continue
		}
		var tg pgtype.UUID
		if err := tg.Scan(tid); err != nil {
			continue
		}
		exists, err := qtx.TagExists(ctx, tg)
		if err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		if !exists {
			_ = tx.Rollback(ctx)
			return ErrParamsType
		}
		if err := qtx.AddPostTag(ctx, dao.AddPostTagParams{
			PostID: pid,
			TagID:  tg,
		}); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) GetPostStatus(ctx context.Context, postID string) (string, error) {
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return "", err
	}
	status, err := r.dao.GetPostStatusByID(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrPostNotExist
		}
		r.log.Error(err)
		return "", ErrDefault
	}
	return status, nil
}
