package repo

import (
	"context"
	"errors"
	postconstant "nurture/internal/post/constant"
	"nurture/internal/post/repo/cache"
	"nurture/internal/post/repo/dao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostRepo) ListHome(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	type listCache struct {
		Rows    []PostRow `json:"rows"`
		HasMore bool      `json:"has_more"`
	}
	key := cache.HotListKey(userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached.Rows, cached.HasMore, nil
		}
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		rows []dao.ListHomeByCtimeRow
		err  error
	)
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "recommend":
		rows, hasMore, err := r.listRecommend(ctx, userID, page, pageSize)
		if err == nil {
			return rows, hasMore, nil
		}
		if !errors.Is(err, ErrParamsType) && !errors.Is(err, ErrPostNotExist) {
			r.log.Error(err)
		}
		fallthrough
	case "hot":
		hotRows, e := r.dao.ListHomeByHot(ctx, dao.ListHomeByHotParams{
			Limit:   limit,
			Offset:  offset,
			Column3: userID,
		})
		rows = make([]dao.ListHomeByCtimeRow, len(hotRows))
		for i := range hotRows {
			rows[i] = dao.ListHomeByCtimeRow(hotRows[i])
		}
		err = e
	case "random":
		seed := time.Now().Format("2006-01-02")
		rndRows, e := r.dao.ListHomeByRandom(ctx, dao.ListHomeByRandomParams{
			Column1: seed,
			Limit:   limit,
			Offset:  offset,
			Column4: userID,
		})
		rows = make([]dao.ListHomeByCtimeRow, len(rndRows))
		for i := range rndRows {
			rows[i] = dao.ListHomeByCtimeRow(rndRows[i])
		}
		err = e
	default:
		rows, err = r.dao.ListHomeByCtime(ctx, dao.ListHomeByCtimeParams{
			Limit:   limit,
			Offset:  offset,
			Column3: userID,
		})
	}
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
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = cache.SetJSON(ctx, r.rdb, key, payload, time.Duration(postconstant.HotListTTL)*time.Second)
	}
	return res, hasMore, nil
}

func (r *PostRepo) ListFollowing(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListFollowingByHot(ctx, dao.ListFollowingByHotParams{
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
	rows, err := r.dao.ListFollowingByCtime(ctx, dao.ListFollowingByCtimeParams{
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

func (r *PostRepo) ListByTag(ctx context.Context, userID, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	type listCache struct {
		Rows    []PostRow `json:"rows"`
		HasMore bool      `json:"has_more"`
	}
	key := cache.HotListByTagKey(userID, tagID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached.Rows, cached.HasMore, nil
		}
	}
	var tg pgtype.UUID
	if err := tg.Scan(tagID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		rowsAny any
		err     error
	)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, e := r.dao.ListPostsByTagHot(ctx, dao.ListPostsByTagHotParams{
			TagID:   tg,
			Limit:   limit,
			Offset:  offset,
			Column4: userID,
		})
		rowsAny = rows
		err = e
	} else {
		rows, e := r.dao.ListPostsByTag(ctx, dao.ListPostsByTagParams{
			TagID:   tg,
			Status:  "published",
			Limit:   limit,
			Offset:  offset,
			Column5: userID,
		})
		rowsAny = rows
		err = e
	}
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]PostRow, 0, pageSize)
	realLen := 0
	switch rows := rowsAny.(type) {
	case []dao.ListPostsByTagHotRow:
		realLen = len(rows)
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
	case []dao.ListPostsByTagRow:
		realLen = len(rows)
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
	default:
		return nil, false, ErrDefault
	}
	hasMore := int32(realLen) >= limit
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = cache.SetJSON(ctx, r.rdb, key, payload, time.Duration(postconstant.HotListTTL)*time.Second)
	}
	return res, hasMore, nil
}

func (r *PostRepo) Search(ctx context.Context, userID, keyword, tagID, strategy string, page, pageSize int) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	kw := "%" + strings.TrimSpace(keyword) + "%"
	var (
		err     error
		resRows []PostRow
		hasMore bool
	)
	if strings.TrimSpace(tagID) != "" {
		var tg pgtype.UUID
		if err := tg.Scan(tagID); err != nil {
			return nil, false, ErrParamsType
		}
		if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
			rows, e := r.dao.SearchPostsByTitleAndTagHot(ctx, dao.SearchPostsByTitleAndTagHotParams{
				Title:   kw,
				TagID:   tg,
				Limit:   limit,
				Offset:  offset,
				Column5: userID,
			})
			err = e
			resRows = make([]PostRow, 0, pageSize)
			for i, v := range rows {
				if int32(i) >= limit-1 {
					break
				}
				resRows = append(resRows, toPostRow(
					v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
					v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
					v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
				))
			}
			hasMore = int32(len(rows)) >= limit
		} else {
			rows, e := r.dao.SearchPostsByTitleAndTagCtime(ctx, dao.SearchPostsByTitleAndTagCtimeParams{
				Title:   kw,
				TagID:   tg,
				Limit:   limit,
				Offset:  offset,
				Column5: userID,
			})
			err = e
			resRows = make([]PostRow, 0, pageSize)
			for i, v := range rows {
				if int32(i) >= limit-1 {
					break
				}
				resRows = append(resRows, toPostRow(
					v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
					v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
					v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
				))
			}
			hasMore = int32(len(rows)) >= limit
		}
	} else {
		if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
			rows, e := r.dao.SearchPostsByTitleHot(ctx, dao.SearchPostsByTitleHotParams{
				Title:   kw,
				Limit:   limit,
				Offset:  offset,
				Column4: userID,
			})
			err = e
			resRows = make([]PostRow, 0, pageSize)
			for i, v := range rows {
				if int32(i) >= limit-1 {
					break
				}
				resRows = append(resRows, toPostRow(
					v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
					v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
					v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
				))
			}
			hasMore = int32(len(rows)) >= limit
		} else {
			rows, e := r.dao.SearchPosts(ctx, dao.SearchPostsParams{
				Title:   kw,
				Status:  "published",
				Limit:   limit,
				Offset:  offset,
				Column5: userID,
			})
			err = e
			resRows = make([]PostRow, 0, pageSize)
			for i, v := range rows {
				if int32(i) >= limit-1 {
					break
				}
				resRows = append(resRows, toPostRow(
					v.PostID, v.AuthorID, v.AuthorName, v.AuthorAvatar, v.AuthorProvince, v.AuthorCity,
					v.Title, v.Content, v.Status, v.LikeCount, v.DislikeCount, v.CollectCount, v.CommentCount,
					v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
				))
			}
			hasMore = int32(len(rows)) >= limit
		}
	}
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	return resRows, hasMore, nil
}

func (r *PostRepo) ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		err  error
		rows []dao.ListPostsByAuthorRow
	)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		hotRows, e := r.dao.ListPostsByAuthorHot(ctx, dao.ListPostsByAuthorHotParams{
			Column1: authorID,
			Limit:   limit,
			Offset:  offset,
		})
		rows = make([]dao.ListPostsByAuthorRow, len(hotRows))
		for i := range hotRows {
			rows[i] = dao.ListPostsByAuthorRow(hotRows[i])
		}
		err = e
	} else {
		rows, err = r.dao.ListPostsByAuthor(ctx, dao.ListPostsByAuthorParams{
			Column1: authorID,
			Limit:   limit,
			Offset:  offset,
		})
	}
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

func (r *PostRepo) ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListDraftsByAuthor(ctx, dao.ListDraftsByAuthorParams{
		Column1: authorID,
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

func (r *PostRepo) ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListMilestonesByAuthor(ctx, dao.ListMilestonesByAuthorParams{
		Column1: authorID,
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
