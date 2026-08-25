package repo

import (
	"context"
	"nurture/internal/post/repo/cache"
	"nurture/internal/post/repo/dao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostRepo) LikePost(ctx context.Context, postID, userID string) error {
	var pid, uid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	status, err := r.dao.GetPostStatusByID(ctx, pid)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if status != "published" {
		return ErrInvalidPostStatus
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreatePostLike(ctx, dao.CreatePostLikeParams{
		UserID: uid,
		PostID: pid,
		Ctime:  now,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncPostLikeCount(ctx, pid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = cache.Del(ctx, r.rdb, cache.HotDetailKey(postID, userID))
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListPattern(userID), 100)
	return nil
}

func (r *PostRepo) UnlikePost(ctx context.Context, postID, userID string) error {
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
	aff, err := qtx.DeletePostLike(ctx, dao.DeletePostLikeParams{
		UserID: uid,
		PostID: pid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecPostLikeCount(ctx, pid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = cache.Del(ctx, r.rdb, cache.HotDetailKey(postID, userID))
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListPattern(userID), 100)
	return nil
}

func (r *PostRepo) CollectPost(ctx context.Context, postID, userID, collectionID string) error {
	var pid, uid, cid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	if err := cid.Scan(collectionID); err != nil {
		return err
	}
	status, err := r.dao.GetPostStatusByID(ctx, pid)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if status != "published" {
		return ErrInvalidPostStatus
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreateCollection(ctx, dao.CreateCollectionParams{
		CollectionID: cid,
		UserID:       uid,
		PostID:       pid,
		Ctime:        now,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncPostCollectCount(ctx, pid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = cache.Del(ctx, r.rdb, cache.HotDetailKey(postID, userID))
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListPattern(userID), 100)
	return nil
}

func (r *PostRepo) UncollectPost(ctx context.Context, postID, userID string) error {
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
	aff, err := qtx.DeleteCollection(ctx, dao.DeleteCollectionParams{
		UserID: uid,
		PostID: pid,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecPostCollectCount(ctx, pid); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = cache.Del(ctx, r.rdb, cache.HotDetailKey(postID, userID))
	_ = cache.ScanDel(ctx, r.rdb, cache.HotListPattern(userID), 100)
	return nil
}

func (r *PostRepo) ListMyCollections(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListCollectionsByHot(ctx, dao.ListCollectionsByHotParams{
			Column1: uid,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			r.log.Error(err)
			return nil, false, ErrDefault
		}
		res := make([]PostRow, 0, pageSize)
		for i, v := range rows {
			if int32(i) >= limit-1 {
				break
			}
			res = append(res, toPostRow(
				v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
				v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
				v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
			))
		}
		hasMore := int32(len(rows)) >= limit
		return res, hasMore, nil
	}
	rows, err := r.dao.ListCollectionsByCtime(ctx, dao.ListCollectionsByCtimeParams{
		Column1: uid,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]PostRow, 0, pageSize)
	for i, v := range rows {
		if int32(i) >= limit-1 {
			break
		}
		res = append(res, toPostRow(
			v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
			v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}
