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
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostRepo) DeleteComment(ctx context.Context, commentID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	var cid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	meta, err := qtx.GetCommentMetaByID(ctx, cid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if meta.Status != "visible" {
		_ = tx.Rollback(ctx)
		return ErrInvalidPostStatus
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	aff, err := qtx.DeleteCommentVisibleByOwner(ctx, dao.DeleteCommentVisibleByOwnerParams{
		CommentID: cid,
		Utime:     time.Now().UnixMilli(),
		UserID:    uid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff == 0 {
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// adjust counters
	if meta.ParentID.Valid {
		if _, err := qtx.DecCommentReplyCount(ctx, meta.ParentID); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	} else {
		var pid pgtype.UUID
		if err := pid.Scan(meta.PostID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if _, err := qtx.DecPostCommentCount(ctx, pid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) UpdateComment(ctx context.Context, commentID, userID, content string) error {
	var cid, uid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	aff, err := r.dao.UpdateCommentContentByOwner(ctx, dao.UpdateCommentContentByOwnerParams{
		CommentID: cid,
		Content:   pgtype.Text{String: content, Valid: true},
		Utime:     time.Now().UnixMilli(),
		UserID:    uid,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrInvalidPostStatus
	}
	return nil
}

func (r *PostRepo) LikeComment(ctx context.Context, commentID, userID string) error {
	var cid, uid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	// check status visible
	meta, err := r.dao.GetCommentMetaByID(ctx, cid)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if meta.Status != "visible" {
		return ErrInvalidPostStatus
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreateCommentLike(ctx, dao.CreateCommentLikeParams{
		UserID:    uid,
		CommentID: cid,
		Ctime:     now,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncCommentLikeCount(ctx, cid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) UnlikeComment(ctx context.Context, commentID, userID string) error {
	var cid, uid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
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
	aff, err := qtx.DeleteCommentLike(ctx, dao.DeleteCommentLikeParams{
		UserID:    uid,
		CommentID: cid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecCommentLikeCount(ctx, cid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) ListRepliesByComment(ctx context.Context, commentID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error) {
	type listCache struct {
		Rows    []CommentRow `json:"rows"`
		HasMore bool         `json:"has_more"`
	}
	key := cache.CommentHotRepliesKey(commentID, userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached.Rows, cached.HasMore, nil
		}
	}
	var cid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListCommentRepliesByHot(ctx, dao.ListCommentRepliesByHotParams{
			ParentID: cid,
			Limit:    limit,
			Offset:   offset,
			Column4:  userID,
		})
		if err != nil {
			r.log.Error(err)
			return nil, false, ErrDefault
		}
		res := make([]CommentRow, 0, pageSize)
		for i, v := range rows {
			if int32(i) >= limit-1 {
				break
			}
			res = append(res, CommentRow{
				CommentID:  v.CommentID,
				UserID:     v.UserID,
				Username:   v.Username,
				Avatar:     v.Avatar,
				Content:    v.Content.String,
				LikeCount:  v.LikeCount,
				ReplyCount: int32(v.ReplyCount),
				Ctime:      v.Ctime,
				Utime:      v.Utime,
			})
		}
		hasMore := int32(len(rows)) >= limit
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = cache.SetJSON(ctx, r.rdb, key, payload, time.Duration(postconstant.CommentHotRepliesTTL)*time.Second)
		return res, hasMore, nil
	}
	rows, err := r.dao.ListCommentRepliesByCtime(ctx, dao.ListCommentRepliesByCtimeParams{
		ParentID: cid,
		Limit:    limit,
		Offset:   offset,
		Column4:  userID,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]CommentRow, 0, pageSize)
	for i, v := range rows {
		if int32(i) >= limit-1 {
			break
		}
		res = append(res, CommentRow{
			CommentID:  v.CommentID,
			UserID:     v.UserID,
			Username:   v.Username,
			Avatar:     v.Avatar,
			Content:    v.Content.String,
			LikeCount:  v.LikeCount,
			ReplyCount: int32(v.ReplyCount),
			Ctime:      v.Ctime,
			Utime:      v.Utime,
		})
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) GetCommentParentInfo(ctx context.Context, commentID string) (string, string, error) {
	var cid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		return "", "", err
	}
	row, err := r.dao.GetCommentMinimal(ctx, cid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrPostNotExist
		}
		r.log.Error(err)
		return "", "", ErrDefault
	}
	// row has post_id and status
	return row.PostID, row.Status, nil
}

func (r *PostRepo) CreateComment(ctx context.Context, commentID, postID, userID string, parentID *string, content string, now int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	var (
		cid, pid, uid pgtype.UUID
		pgid          pgtype.UUID
	)
	if err := cid.Scan(commentID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := pid.Scan(postID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := uid.Scan(userID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	var parent pgtype.UUID
	if parentID != nil && strings.TrimSpace(*parentID) != "" {
		if err := pgid.Scan(*parentID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		parent = pgid
	}
	if err := qtx.CreateComment(ctx, dao.CreateCommentParams{
		CommentID: cid,
		PostID:    pid,
		UserID:    uid,
		ParentID:  parent,
		Content:   pgtype.Text{String: content, Valid: true},
		Ctime:     now,
		Utime:     now,
	}); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if _, err := qtx.IncPostCommentCount(ctx, pid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// if has parent, inc reply_count
	if parentID != nil && strings.TrimSpace(*parentID) != "" {
		if _, err := qtx.IncCommentReplyCount(ctx, pgid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// 缓存失效与重建
	_ = cache.Del(ctx, r.rdb, cache.HotDetailKey(postID, userID))
	_ = cache.ScanDel(ctx, r.rdb, cache.HotCommentsPattern(postID, userID), 100)
	if parentID != nil && strings.TrimSpace(*parentID) != "" {
		_ = cache.ScanDel(ctx, r.rdb, cache.CommentHotRepliesPattern(*parentID, userID), 100)
	}
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListPattern(userID), 100)
	_, _ = r.GetDetail(ctx, userID, postID)
	return nil
}

type CommentRow struct {
	CommentID  string
	UserID     string
	Username   string
	Avatar     string
	Content    string
	LikeCount  int32
	ReplyCount int32
	Ctime      int64
	Utime      int64
	HasLiked   bool
}

func (r *PostRepo) ListCommentsByPost(ctx context.Context, postID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error) {
	type listCache struct {
		Rows    []CommentRow `json:"rows"`
		HasMore bool         `json:"has_more"`
	}
	key := cache.HotCommentsKey(postID, userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached.Rows, cached.HasMore, nil
		}
	}
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListPostCommentsByHot(ctx, dao.ListPostCommentsByHotParams{
			PostID:  pid,
			Limit:   limit,
			Offset:  offset,
			Column4: userID,
		})
		if err != nil {
			r.log.Error(err)
			return nil, false, ErrDefault
		}
		res := make([]CommentRow, 0, pageSize)
		for i, v := range rows {
			if int32(i) >= limit-1 {
				break
			}
			res = append(res, CommentRow{
				CommentID:  v.CommentID,
				UserID:     v.UserID,
				Username:   v.Username,
				Avatar:     v.Avatar,
				Content:    v.Content.String,
				LikeCount:  v.LikeCount,
				ReplyCount: v.ReplyCount,
				Ctime:      v.Ctime,
				Utime:      v.Utime,
				HasLiked:   v.HasLiked,
			})
		}
		hasMore := int32(len(rows)) >= limit
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = cache.SetJSON(ctx, r.rdb, key, payload, time.Duration(postconstant.HotCommentsTTL)*time.Second)
		return res, hasMore, nil
	}
	rows, err := r.dao.ListPostCommentsByCtime(ctx, dao.ListPostCommentsByCtimeParams{
		PostID:  pid,
		Limit:   limit,
		Offset:  offset,
		Column4: userID,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]CommentRow, 0, pageSize)
	for i, v := range rows {
		if int32(i) >= limit-1 {
			break
		}
		res = append(res, CommentRow{
			CommentID:  v.CommentID,
			UserID:     v.UserID,
			Username:   v.Username,
			Avatar:     v.Avatar,
			Content:    v.Content.String,
			LikeCount:  v.LikeCount,
			ReplyCount: v.ReplyCount,
			Ctime:      v.Ctime,
			Utime:      v.Utime,
			HasLiked:   v.HasLiked,
		})
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}
