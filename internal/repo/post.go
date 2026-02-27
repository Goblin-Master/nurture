package repo

import (
	"context"
	"errors"
	"nurture/internal/global"
	"nurture/internal/repo/post"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostRow struct {
	PostID         string
	AuthorID       string
	AuthorName     string
	AuthorAvatar   string
	AuthorProvince string
	AuthorCity     string
	Title          string
	Content        string
	Status         string
	LikeCount      int32
	DislikeCount   int32
	CollectCount   int32
	CommentCount   int32
	Ctime          int64
	Utime          int64
	Birthday       int64
	Tags           []string
}

type IPostRepo interface {
	ListHome(ctx context.Context, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListByTag(ctx context.Context, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	Search(ctx context.Context, keyword, tagID, strategy string, page, pageSize int) ([]PostRow, bool, error)
	ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListFollowing(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	GetDetail(ctx context.Context, postID string) (PostRow, error)
	CreatePost(ctx context.Context, postID, authorID, title, content, status string, ctime, utime int64, tagIDs []string) error
	Publish(ctx context.Context, postID, userID string) error
	UpdateDraft(ctx context.Context, postID, userID, title, content string, tagIDs []string) error
	CreateComment(ctx context.Context, commentID, postID, userID string, parentID *string, content string, now int64) error
	GetPostStatus(ctx context.Context, postID string) (string, error)
	GetCommentParentInfo(ctx context.Context, commentID string) (string, string, error)
	ListCommentsByPost(ctx context.Context, postID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error)
	ListRepliesByComment(ctx context.Context, commentID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error)
	DeleteComment(ctx context.Context, commentID, userID string) error
	UpdateComment(ctx context.Context, commentID, userID, content string) error
	LikeComment(ctx context.Context, commentID, userID string) error
	UnlikeComment(ctx context.Context, commentID, userID string) error
}

type PostRepo struct {
	dao *post.Queries
}

func NewPostRepo() *PostRepo {
	return &PostRepo{
		dao: post.New(global.DB),
	}
}

var _ IPostRepo = (*PostRepo)(nil)

// 已删除旧的动态 List 方法

func (r *PostRepo) GetDetail(ctx context.Context, postID string) (PostRow, error) {
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return PostRow{}, err
	}
	row, err := r.dao.GetPostDetail(ctx, pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PostRow{}, ErrPostNotExist
		}
		global.Log.Error(err)
		return PostRow{}, ErrDefault
	}
	var tags []string
	switch v := row.Tags.(type) {
	case []string:
		tags = v
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
	return PostRow{
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
	}, nil
}

func toPostRow(postID, authorID, authorName, authorAvatar, authorProvince, authorCity, title, content, status string,
	likeCount, dislikeCount, collectCount, commentCount int32,
	ctime, utime int64, birthday interface{}, tags interface{}) PostRow {
	var tagList []string
	switch v := tags.(type) {
	case []string:
		tagList = v
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
	}
}

func (r *PostRepo) ListHome(ctx context.Context, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		rows []post.ListHomeByCtimeRow
		err  error
	)
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "hot":
		hotRows, e := r.dao.ListHomeByHot(ctx, post.ListHomeByHotParams{
			Limit:  limit,
			Offset: offset,
		})
		rows = make([]post.ListHomeByCtimeRow, len(hotRows))
		for i := range hotRows {
			rows[i] = post.ListHomeByCtimeRow(hotRows[i])
		}
		err = e
	case "random":
		seed := time.Now().Format("2006-01-02")
		rndRows, e := r.dao.ListHomeByRandom(ctx, post.ListHomeByRandomParams{
			Column1: seed,
			Limit:   limit,
			Offset:  offset,
		})
		rows = make([]post.ListHomeByCtimeRow, len(rndRows))
		for i := range rndRows {
			rows[i] = post.ListHomeByCtimeRow(rndRows[i])
		}
		err = e
	default:
		rows, err = r.dao.ListHomeByCtime(ctx, post.ListHomeByCtimeParams{
			Limit:  limit,
			Offset: offset,
		})
	}
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
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
		rows, err := r.dao.ListFollowingByHot(ctx, post.ListFollowingByHotParams{
			Follower: uid,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			global.Log.Error(err)
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
				v.Ctime, v.Utime, v.Birthday, v.Tags,
			))
		}
		hasMore := int32(len(rows)) >= limit
		return res, hasMore, nil
	}
	rows, err := r.dao.ListFollowingByCtime(ctx, post.ListFollowingByCtimeParams{
		Follower: uid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) DeleteComment(ctx context.Context, commentID, userID string) error {
	tx, err := global.DB.Begin(ctx)
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
		global.Log.Error(err)
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
	aff, err := qtx.DeleteCommentVisibleByOwner(ctx, post.DeleteCommentVisibleByOwnerParams{
		CommentID: cid,
		Utime:     time.Now().UnixMilli(),
		UserID:    uid,
	})
	if err != nil {
		global.Log.Error(err)
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
			global.Log.Error(err)
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
			global.Log.Error(err)
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
	aff, err := r.dao.UpdateCommentContentByOwner(ctx, post.UpdateCommentContentByOwnerParams{
		CommentID: cid,
		Content:   pgtype.Text{String: content, Valid: true},
		Utime:     time.Now().UnixMilli(),
		UserID:    uid,
	})
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
		return ErrDefault
	}
	if meta.Status != "visible" {
		return ErrInvalidPostStatus
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreateCommentLike(ctx, post.CreateCommentLikeParams{
		UserID:    uid,
		CommentID: cid,
		Ctime:     now,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncCommentLikeCount(ctx, cid); err != nil {
			global.Log.Error(err)
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
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	aff, err := qtx.DeleteCommentLike(ctx, post.DeleteCommentLikeParams{
		UserID:    uid,
		CommentID: cid,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecCommentLikeCount(ctx, cid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}
func (r *PostRepo) ListRepliesByComment(ctx context.Context, commentID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error) {
	var cid pgtype.UUID
	if err := cid.Scan(commentID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListCommentRepliesByHot(ctx, post.ListCommentRepliesByHotParams{
			ParentID: cid,
			Limit:    limit,
			Offset:   offset,
			Column4:  userID,
		})
		if err != nil {
			global.Log.Error(err)
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
	rows, err := r.dao.ListCommentRepliesByCtime(ctx, post.ListCommentRepliesByCtimeParams{
		ParentID: cid,
		Limit:    limit,
		Offset:   offset,
		Column4:  userID,
	})
	if err != nil {
		global.Log.Error(err)
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

func (r *PostRepo) ListByTag(ctx context.Context, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	var tg pgtype.UUID
	if err := tg.Scan(tagID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		rows []post.ListPostsByTagRow
		err  error
	)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		hotRows, e := r.dao.ListPostsByTagHot(ctx, post.ListPostsByTagHotParams{
			TagID:  tg,
			Limit:  limit,
			Offset: offset,
		})
		rows = make([]post.ListPostsByTagRow, len(hotRows))
		for i := range hotRows {
			rows[i] = post.ListPostsByTagRow(hotRows[i])
		}
		err = e
	} else {
		rows, err = r.dao.ListPostsByTag(ctx, post.ListPostsByTagParams{
			TagID:  tg,
			Status: "published",
			Limit:  limit,
			Offset: offset,
		})
	}
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) Search(ctx context.Context, keyword, tagID, strategy string, page, pageSize int) ([]PostRow, bool, error) {
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
			rows, e := r.dao.SearchPostsByTitleAndTagHot(ctx, post.SearchPostsByTitleAndTagHotParams{
				Title:  kw,
				TagID:  tg,
				Limit:  limit,
				Offset: offset,
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
					v.Ctime, v.Utime, v.Birthday, v.Tags,
				))
			}
			hasMore = int32(len(rows)) >= limit
		} else {
			rows, e := r.dao.SearchPostsByTitleAndTagCtime(ctx, post.SearchPostsByTitleAndTagCtimeParams{
				Title:  kw,
				TagID:  tg,
				Limit:  limit,
				Offset: offset,
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
					v.Ctime, v.Utime, v.Birthday, v.Tags,
				))
			}
			hasMore = int32(len(rows)) >= limit
		}
	} else {
		if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
			rows, e := r.dao.SearchPostsByTitleHot(ctx, post.SearchPostsByTitleHotParams{
				Title:  kw,
				Limit:  limit,
				Offset: offset,
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
					v.Ctime, v.Utime, v.Birthday, v.Tags,
				))
			}
			hasMore = int32(len(rows)) >= limit
		} else {
			rows, e := r.dao.SearchPosts(ctx, post.SearchPostsParams{
				Title:  kw,
				Status: "published",
				Limit:  limit,
				Offset: offset,
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
					v.Ctime, v.Utime, v.Birthday, v.Tags,
				))
			}
			hasMore = int32(len(rows)) >= limit
		}
	}
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	return resRows, hasMore, nil
}

func (r *PostRepo) ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	var aid pgtype.UUID
	if err := aid.Scan(authorID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		err  error
		rows []post.ListPostsByAuthorRow
	)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		hotRows, e := r.dao.ListPostsByAuthorHot(ctx, post.ListPostsByAuthorHotParams{
			AuthorID: aid,
			Limit:    limit,
			Offset:   offset,
		})
		rows = make([]post.ListPostsByAuthorRow, len(hotRows))
		for i := range hotRows {
			rows[i] = post.ListPostsByAuthorRow(hotRows[i])
		}
		err = e
	} else {
		rows, err = r.dao.ListPostsByAuthor(ctx, post.ListPostsByAuthorParams{
			AuthorID: aid,
			Limit:    limit,
			Offset:   offset,
		})
	}
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	var aid pgtype.UUID
	if err := aid.Scan(authorID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListDraftsByAuthor(ctx, post.ListDraftsByAuthorParams{
		AuthorID: aid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	var aid pgtype.UUID
	if err := aid.Scan(authorID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListMilestonesByAuthor(ctx, post.ListMilestonesByAuthorParams{
		AuthorID: aid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		global.Log.Error(err)
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
			v.Ctime, v.Utime, v.Birthday, v.Tags,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}
func (r *PostRepo) CreatePost(ctx context.Context, postID, authorID, title, content, status string, ctime, utime int64, tagIDs []string) error {
	var pid, aid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := aid.Scan(authorID); err != nil {
		return err
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	err = qtx.CreatePost(ctx, post.CreatePostParams{
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
		global.Log.Error(err)
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
		if err := qtx.AddPostTag(ctx, post.AddPostTagParams{
			PostID: pid,
			TagID:  tg,
		}); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) Publish(ctx context.Context, postID, userID string) error {
	var pid, uid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return err
	}
	if err := uid.Scan(userID); err != nil {
		return err
	}
	count, err := r.dao.PublishPost(ctx, post.PublishPostParams{
		PostID:   pid,
		AuthorID: uid,
		Utime:    time.Now().UnixMilli(),
	})
	if err != nil {
		global.Log.Error(err)
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
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	aff, err := qtx.UpdateDraftByOwner(ctx, post.UpdateDraftByOwnerParams{
		PostID:   pid,
		AuthorID: uid,
		Title:    title,
		Content:  content,
		Utime:    time.Now().UnixMilli(),
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff == 0 {
		_ = tx.Rollback(ctx)
		return ErrPostNotDraft
	}
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		global.Log.Error(err)
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
		if err := qtx.AddPostTag(ctx, post.AddPostTagParams{
			PostID: pid,
			TagID:  tg,
		}); err != nil {
			global.Log.Error(err)
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
		global.Log.Error(err)
		return "", ErrDefault
	}
	return status, nil
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
		global.Log.Error(err)
		return "", "", ErrDefault
	}
	// row has post_id and status
	return row.PostID, row.Status, nil
}

func (r *PostRepo) CreateComment(ctx context.Context, commentID, postID, userID string, parentID *string, content string, now int64) error {
	tx, err := global.DB.Begin(ctx)
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
	if err := qtx.CreateComment(ctx, post.CreateCommentParams{
		CommentID: cid,
		PostID:    pid,
		UserID:    uid,
		ParentID:  parent,
		Content:   pgtype.Text{String: content, Valid: true},
		Ctime:     now,
		Utime:     now,
	}); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if _, err := qtx.IncPostCommentCount(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// if has parent, inc reply_count
	if parentID != nil && strings.TrimSpace(*parentID) != "" {
		if _, err := qtx.IncCommentReplyCount(ctx, pgid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	return tx.Commit(ctx)
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
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return nil, false, ErrParamsType
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		rows, err := r.dao.ListPostCommentsByHot(ctx, post.ListPostCommentsByHotParams{
			PostID:  pid,
			Limit:   limit,
			Offset:  offset,
			Column4: userID,
		})
		if err != nil {
			global.Log.Error(err)
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
	rows, err := r.dao.ListPostCommentsByCtime(ctx, post.ListPostCommentsByCtimeParams{
		PostID:  pid,
		Limit:   limit,
		Offset:  offset,
		Column4: userID,
	})
	if err != nil {
		global.Log.Error(err)
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
