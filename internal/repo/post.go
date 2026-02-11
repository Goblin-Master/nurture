package repo

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/global"
	"nurture/internal/repo/postdao"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type PostListFilter struct {
	Page       int
	PageSize   int
	Status     string
	TagID      string
	AuthorID   string
	OrderBy    string
	Order      string
	Keyword    string
	Strategy   string
	ExcludeIDs []string
}

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
	Cover          string
	Ctime          int64
	Utime          int64
	Birthday       int64
	Tags           []string
}

type IPostRepo interface {
	List(ctx context.Context, f PostListFilter) ([]PostRow, bool, error)
	GetDetail(ctx context.Context, postID string) (PostRow, error)
	CreatePost(ctx context.Context, postID, authorID, title, content, status, cover string, ctime, utime int64, tagIDs []string) error
	Publish(ctx context.Context, postID, userID string) error
}

type PostRepo struct {
	dao *postdao.Queries
}

func NewPostRepo() *PostRepo {
	return &PostRepo{
		dao: postdao.New(global.DB),
	}
}

var _ IPostRepo = (*PostRepo)(nil)

func (r *PostRepo) List(ctx context.Context, f PostListFilter) ([]PostRow, bool, error) {
	limit := f.PageSize + 1
	offset := (f.Page - 1) * f.PageSize

	sb := strings.Builder{}
	sb.WriteString(`SELECT 
    p.post_id::text AS post_id,
    p.author_id::text AS author_id,
    COALESCE(ua.username, '') AS author_name,
    COALESCE(ua.avatar, '') AS author_avatar,
    COALESCE(ua.province, '') AS author_province,
    COALESCE(ua.city, '') AS author_city,
    p.title, p.content, p.status,
    p.like_count, p.dislike_count, p.collect_count, p.comment_count,
    p.cover, p.ctime, p.utime,
    COALESCE(b.birthday, 0) AS birthday,
    COALESCE(array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL), '{}') AS tags
  FROM "post" p
  LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
  LEFT JOIN LATERAL (
    SELECT birthday FROM "baby" 
    WHERE user_id = p.author_id 
    ORDER BY ctime DESC 
    LIMIT 1
  ) b ON true
  LEFT JOIN "post_tag" pt ON pt.post_id = p.post_id
  LEFT JOIN "tag" t ON t.tag_id = pt.tag_id`)

	conds := make([]string, 0, 6)
	args := make([]any, 0, 6)
	argIdx := 1

	// 标签过滤
	if f.TagID != "" {
		conds = append(conds, fmt.Sprintf("pt.tag_id = $%d", argIdx))
		args = append(args, f.TagID)
		argIdx++
	}
	// 状态过滤（默认 published）
	status := f.Status
	if status == "" {
		status = "published"
	}
	conds = append(conds, fmt.Sprintf(`status = $%d`, argIdx))
	args = append(args, status)
	argIdx++

	// 作者过滤
	if f.AuthorID != "" {
		conds = append(conds, fmt.Sprintf(`author_id = $%d`, argIdx))
		args = append(args, f.AuthorID)
		argIdx++
	}

	// 关键字（标题或正文 LIKE）
	if f.Keyword != "" {
		conds = append(conds, fmt.Sprintf(`(p.title ILIKE $%d OR COALESCE(ua.username,'') ILIKE $%d)`, argIdx, argIdx))
		args = append(args, "%"+f.Keyword+"%")
		argIdx++
	}
	// 排除已展示
	if len(f.ExcludeIDs) > 0 {
		conds = append(conds, fmt.Sprintf(`NOT (p.post_id::text = ANY($%d::text[]))`, argIdx))
		args = append(args, f.ExcludeIDs)
		argIdx++
	}

	if len(conds) > 0 {
		sb.WriteString(` WHERE ` + strings.Join(conds, " AND "))
	}
	sb.WriteString(` GROUP BY 
    p.post_id, p.author_id, ua.username, ua.avatar, ua.province, ua.city,
    p.title, p.content, p.status, p.like_count, p.dislike_count, p.collect_count, p.comment_count,
    p.cover, p.ctime, p.utime, b.birthday`)

	// 排序
	if f.Strategy == "hot" {
		sb.WriteString(` ORDER BY (like_count*3 + comment_count*5 + collect_count*4) DESC, ctime DESC`)
	} else {
		orderBy := f.OrderBy
		switch orderBy {
		case "ctime", "like_count", "comment_count":
		default:
			orderBy = "ctime"
		}
		order := strings.ToUpper(f.Order)
		if order != "ASC" {
			order = "DESC"
		}
		sb.WriteString(fmt.Sprintf(` ORDER BY %s %s`, orderBy, order))
	}
	sb.WriteString(fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1))
	args = append(args, limit, offset)

	rows, err := global.DB.Query(ctx, sb.String(), args...)
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	defer rows.Close()

	res := make([]PostRow, 0, f.PageSize)
	var count int
	for rows.Next() {
		var rrow PostRow
		if err := rows.Scan(
			&rrow.PostID,
			&rrow.AuthorID,
			&rrow.AuthorName,
			&rrow.AuthorAvatar,
			&rrow.AuthorProvince,
			&rrow.AuthorCity,
			&rrow.Title,
			&rrow.Content,
			&rrow.Status,
			&rrow.LikeCount,
			&rrow.DislikeCount,
			&rrow.CollectCount,
			&rrow.CommentCount,
			&rrow.Cover,
			&rrow.Ctime,
			&rrow.Utime,
			&rrow.Birthday,
			&rrow.Tags,
		); err != nil {
			global.Log.Error(err)
			return nil, false, ErrDefault
		}
		count++
		if count <= f.PageSize {
			res = append(res, rrow)
		}
	}
	hasMore := count > f.PageSize
	return res, hasMore, nil
}

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
		Cover:          row.Cover,
		Ctime:          row.Ctime,
		Utime:          row.Utime,
		Birthday:       row.Birthday,
		Tags:           tags,
	}, nil
}

func (r *PostRepo) CreatePost(ctx context.Context, postID, authorID, title, content, status, cover string, ctime, utime int64, tagIDs []string) error {
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
	err = qtx.CreatePost(ctx, postdao.CreatePostParams{
		PostID:   pid,
		AuthorID: aid,
		Title:    title,
		Content:  content,
		Status:   status,
		Cover:    cover,
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
		if err := qtx.AddPostTag(ctx, postdao.AddPostTagParams{
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
	cmd, err := global.DB.Exec(ctx, `
		UPDATE "post"
		SET status = 'published', utime = $3
		WHERE post_id = $1 AND author_id = $2 AND status = 'draft'
	`, pid, uid, time.Now().UnixMilli())
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if cmd.RowsAffected() == 0 {
		return ErrPostNotDraft
	}
	return nil
}
