package repo

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/constant"
	"nurture/internal/global"
	"nurture/internal/repo/post"
	"nurture/internal/repo/user"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tmc/langchaingo/schema"
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
	IsLike         bool
	IsDislike      bool
	IsCollect      bool
}

type TagRow struct {
	TagID       string
	Name        string
	Description string
}

type IPostRepo interface {
	ListHome(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListByTag(ctx context.Context, userID, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	Search(ctx context.Context, userID, keyword, tagID, strategy string, page, pageSize int) ([]PostRow, bool, error)
	ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListFollowing(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	GetDetail(ctx context.Context, userID, postID string) (PostRow, error)
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
	LikePost(ctx context.Context, postID, userID string) error
	UnlikePost(ctx context.Context, postID, userID string) error
	LikeComment(ctx context.Context, commentID, userID string) error
	UnlikeComment(ctx context.Context, commentID, userID string) error
	CollectPost(ctx context.Context, postID, userID, collectionID string) error
	UncollectPost(ctx context.Context, postID, userID string) error
	ListMyCollections(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	// admin tag
	CreateTag(ctx context.Context, tagID, name, description string, now int64) (TagRow, error)
	DeleteTag(ctx context.Context, tagID string) error
	ListTags(ctx context.Context, keyword string, page, pageSize int) ([]TagRow, bool, error)
}

type PostRepo struct {
	dao *post.Queries
	rdb redis.Cmdable
}

func NewPostRepo() *PostRepo {
	return &PostRepo{
		dao: post.New(global.DB),
		rdb: global.RDB,
	}
}

var _ IPostRepo = (*PostRepo)(nil)

func (r *PostRepo) CreateTag(ctx context.Context, tagID, name, description string, now int64) (TagRow, error) {
	var tid pgtype.UUID
	if err := tid.Scan(tagID); err != nil {
		return TagRow{}, ErrParamsType
	}
	row, err := r.dao.CreateTag(ctx, post.CreateTagParams{
		TagID:       tid,
		TagName:     name,
		Description: pgtype.Text{String: description, Valid: true},
		Ctime:       now,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return TagRow{}, ErrDefault
		}
		global.Log.Error(err)
		return TagRow{}, ErrDefault
	}
	return TagRow{TagID: row.TagID, Name: row.TagName, Description: row.Description}, nil
}

func (r *PostRepo) DeleteTag(ctx context.Context, tagID string) error {
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	var tid pgtype.UUID
	if err := tid.Scan(tagID); err != nil {
		_ = tx.Rollback(ctx)
		return ErrParamsType
	}
	if _, err := qtx.DeletePostTagsByTagID(ctx, tid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if _, err := qtx.DeleteTagByID(ctx, tid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) ListTags(ctx context.Context, keyword string, page, pageSize int) ([]TagRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListTags(ctx, post.ListTagsParams{
		Column1: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	hasMore := int32(len(rows)) >= limit
	res := make([]TagRow, 0, pageSize)
	for i, v := range rows {
		if int32(i) >= limit-1 {
			break
		}
		res = append(res, TagRow{
			TagID:       v.TagID,
			Name:        v.TagName,
			Description: v.Description,
		})
	}
	return res, hasMore, nil
}

func (r *PostRepo) GetDetail(ctx context.Context, userID, postID string) (PostRow, error) {
	key := post.CacheHotDetailKey(postID, userID)
	{
		var cached PostRow
		if ok, _ := getCacheJSON(ctx, r.rdb, key, &cached); ok {
			return cached, nil
		}
	}
	var pid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return PostRow{}, err
	}
	row, err := r.dao.GetPostDetail(ctx, post.GetPostDetailParams{
		PostID:  pid,
		Column2: userID,
	})
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
	_ = setCacheJSON(ctx, r.rdb, key, ret, time.Duration(constant.POST_HOT_DETAIL_TTL)*time.Second)
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

func (r *PostRepo) ListHome(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	type listCache struct {
		Rows    []PostRow `json:"rows"`
		HasMore bool      `json:"has_more"`
	}
	key := post.CacheHotListKey(userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := getCacheJSON(ctx, r.rdb, key, &cached); ok {
			return cached.Rows, cached.HasMore, nil
		}
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		rows []post.ListHomeByCtimeRow
		err  error
	)
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "recommend":
		rows, hasMore, err := r.listRecommend(ctx, userID, page, pageSize)
		if err == nil {
			return rows, hasMore, nil
		}
		if !errors.Is(err, ErrParamsType) && !errors.Is(err, ErrPostNotExist) {
			global.Log.Error(err)
		}
		fallthrough
	case "hot":
		hotRows, e := r.dao.ListHomeByHot(ctx, post.ListHomeByHotParams{
			Limit:   limit,
			Offset:  offset,
			Column3: userID,
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
			Column4: userID,
		})
		rows = make([]post.ListHomeByCtimeRow, len(rndRows))
		for i := range rndRows {
			rows[i] = post.ListHomeByCtimeRow(rndRows[i])
		}
		err = e
	default:
		rows, err = r.dao.ListHomeByCtime(ctx, post.ListHomeByCtimeParams{
			Limit:   limit,
			Offset:  offset,
			Column3: userID,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	hasMore := int32(len(rows)) >= limit
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = setCacheJSON(ctx, r.rdb, key, payload, time.Duration(constant.POST_HOT_LIST_TTL)*time.Second)
	}
	return res, hasMore, nil
}

// listRecommend 推荐策略
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
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	profileText = strings.TrimSpace(profileText)
	if profileText == "" {
		return nil, false, ErrParamsType
	}
	if global.AIX == nil {
		return nil, false, ErrParamsType
	}
	topK := 200
	docs, err := global.AIX.SimilaritySearch(ctx, profileText, []string{constant.COLLECTION_POST_RECOMMEND}, topK)
	if err != nil {
		global.Log.Error(err)
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
	rows, err := r.dao.ListPostsByIDs(ctx, post.ListPostsByIDsParams{
		Column1: arr,
		Column2: userID,
	})
	if err != nil {
		global.Log.Error(err)
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

// TouchUserRecommendProfile 触发用户推荐画像更新
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
		global.Log.Error(err)
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
	err = r.dao.UpsertUserRecommendProfile(ctx, post.UpsertUserRecommendProfileParams{
		UserID:      uid,
		ProfileText: snippet,
		Utime:       now,
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *PostRepo) IndexPostForRecommend(ctx context.Context, postID string) error {
	if global.AIX == nil {
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
		global.Log.Error(err)
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
	err = global.AIX.AddDocumentWithMetadata(ctx, constant.COLLECTION_POST_RECOMMEND, doc, map[string]any{
		"post_id": row.PostID,
	}, false)
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
		return ErrDefault
	}
	if len(tags) == 0 {
		return nil
	}
	key := user.CacheTagPrefKey(userID)
	pipe := r.rdb.TxPipeline()
	for _, t := range tags {
		name := strings.TrimSpace(t)
		if name == "" {
			continue
		}
		pipe.ZIncrBy(ctx, key, score, name)
	}
	pipe.Expire(ctx, key, time.Duration(constant.USER_TAG_PREF_TTL)*time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
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
			Column1: uid,
			Limit:   limit,
			Offset:  offset,
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
				v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
			))
		}
		hasMore := int32(len(rows)) >= limit
		return res, hasMore, nil
	}
	rows, err := r.dao.ListFollowingByCtime(ctx, post.ListFollowingByCtimeParams{
		Column1: uid,
		Limit:   limit,
		Offset:  offset,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
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
		global.Log.Error(err)
		return ErrDefault
	}
	if status != "published" {
		return ErrInvalidPostStatus
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreatePostLike(ctx, post.CreatePostLikeParams{
		UserID: uid,
		PostID: pid,
		Ctime:  now,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncPostLikeCount(ctx, pid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = delCache(ctx, r.rdb, post.CacheHotDetailKey(postID, userID))
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListPattern(userID), 100)
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
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	aff, err := qtx.DeletePostLike(ctx, post.DeletePostLikeParams{
		UserID: uid,
		PostID: pid,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecPostLikeCount(ctx, pid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = delCache(ctx, r.rdb, post.CacheHotDetailKey(postID, userID))
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListPattern(userID), 100)
	return nil
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
	type listCache struct {
		Rows    []CommentRow `json:"rows"`
		HasMore bool         `json:"has_more"`
	}
	key := post.CacheCommentHotRepliesKey(commentID, userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := getCacheJSON(ctx, r.rdb, key, &cached); ok {
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
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = setCacheJSON(ctx, r.rdb, key, payload, time.Duration(constant.COMMENT_HOT_REPLIES_TTL)*time.Second)
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

func (r *PostRepo) ListByTag(ctx context.Context, userID, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	type listCache struct {
		Rows    []PostRow `json:"rows"`
		HasMore bool      `json:"has_more"`
	}
	key := post.CacheHotListByTagKey(userID, tagID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := getCacheJSON(ctx, r.rdb, key, &cached); ok {
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
		rows, e := r.dao.ListPostsByTagHot(ctx, post.ListPostsByTagHotParams{
			TagID:   tg,
			Limit:   limit,
			Offset:  offset,
			Column4: userID,
		})
		rowsAny = rows
		err = e
	} else {
		rows, e := r.dao.ListPostsByTag(ctx, post.ListPostsByTagParams{
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
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	res := make([]PostRow, 0, pageSize)
	realLen := 0
	switch rows := rowsAny.(type) {
	case []post.ListPostsByTagHotRow:
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
	case []post.ListPostsByTagRow:
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
		_ = setCacheJSON(ctx, r.rdb, key, payload, time.Duration(constant.POST_HOT_LIST_TTL)*time.Second)
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
			rows, e := r.dao.SearchPostsByTitleAndTagHot(ctx, post.SearchPostsByTitleAndTagHotParams{
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
			rows, e := r.dao.SearchPostsByTitleAndTagCtime(ctx, post.SearchPostsByTitleAndTagCtimeParams{
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
			rows, e := r.dao.SearchPostsByTitleHot(ctx, post.SearchPostsByTitleHotParams{
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
			rows, e := r.dao.SearchPosts(ctx, post.SearchPostsParams{
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
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	return resRows, hasMore, nil
}

func (r *PostRepo) ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	var (
		err  error
		rows []post.ListPostsByAuthorRow
	)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		hotRows, e := r.dao.ListPostsByAuthorHot(ctx, post.ListPostsByAuthorHotParams{
			Column1: authorID,
			Limit:   limit,
			Offset:  offset,
		})
		rows = make([]post.ListPostsByAuthorRow, len(hotRows))
		for i := range hotRows {
			rows[i] = post.ListPostsByAuthorRow(hotRows[i])
		}
		err = e
	} else {
		rows, err = r.dao.ListPostsByAuthor(ctx, post.ListPostsByAuthorParams{
			Column1: authorID,
			Limit:   limit,
			Offset:  offset,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListDraftsByAuthor(ctx, post.ListDraftsByAuthorParams{
		Column1: authorID,
		Limit:   limit,
		Offset:  offset,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
}

func (r *PostRepo) ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListMilestonesByAuthor(ctx, post.ListMilestonesByAuthorParams{
		Column1: authorID,
		Limit:   limit,
		Offset:  offset,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
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
		exists, err := qtx.TagExists(ctx, tg)
		if err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		if !exists {
			_ = tx.Rollback(ctx)
			return ErrParamsType
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
		global.Log.Error(err)
		return ErrDefault
	}
	if status != "published" {
		return ErrInvalidPostStatus
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	now := time.Now().UnixMilli()
	aff, err := qtx.CreateCollection(ctx, post.CreateCollectionParams{
		CollectionID: cid,
		UserID:       uid,
		PostID:       pid,
		Ctime:        now,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.IncPostCollectCount(ctx, pid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = delCache(ctx, r.rdb, post.CacheHotDetailKey(postID, userID))
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListPattern(userID), 100)
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
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	aff, err := qtx.DeleteCollection(ctx, post.DeleteCollectionParams{
		UserID: uid,
		PostID: pid,
	})
	if err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if aff > 0 {
		if _, err := qtx.DecPostCollectCount(ctx, pid); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = delCache(ctx, r.rdb, post.CacheHotDetailKey(postID, userID))
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListPattern(userID), 100)
	return nil
}

func (r *PostRepo) DeleteDraft(ctx context.Context, postID, authorID string) error {
	var pid, aid pgtype.UUID
	if err := pid.Scan(postID); err != nil {
		return ErrParamsType
	}
	if err := aid.Scan(authorID); err != nil {
		return ErrParamsType
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	aff, err := qtx.DeleteDraftByOwner(ctx, post.DeleteDraftByOwnerParams{
		PostID:   pid,
		AuthorID: aid,
	})
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
		return ErrDefault
	}
	if status == "draft" {
		return ErrInvalidPostStatus
	}
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	if err := qtx.DeleteCommentLikesByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCommentClosuresByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCommentsByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeletePostLikesByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeleteCollectionsByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if err := qtx.DeletePostTagsByPost(ctx, pid); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	aff, err := qtx.DeletePostByOwner(ctx, post.DeletePostByOwnerParams{
		PostID:   pid,
		AuthorID: aid,
	})
	if err != nil {
		global.Log.Error(err)
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
	_ = scanDelCache(ctx, r.rdb, post.CacheHotDetailPattern(postID), 200)
	_ = scanDelCache(ctx, r.rdb, post.CacheHotCommentsAllUsersPattern(postID), 200)
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListAllPattern(), 500)
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListByTagAllPattern(), 500)
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
		rows, err := r.dao.ListCollectionsByHot(ctx, post.ListCollectionsByHotParams{
			Column1: uid,
			Limit:   limit,
			Offset:  offset,
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
				v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
			))
		}
		hasMore := int32(len(rows)) >= limit
		return res, hasMore, nil
	}
	rows, err := r.dao.ListCollectionsByCtime(ctx, post.ListCollectionsByCtimeParams{
		Column1: uid,
		Limit:   limit,
		Offset:  offset,
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
			v.Ctime, v.Utime, v.Birthday, v.Tags, v.IsLike, v.IsDislike, v.IsCollect,
		))
	}
	hasMore := int32(len(rows)) >= limit
	return res, hasMore, nil
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
		exists, err := qtx.TagExists(ctx, tg)
		if err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		if !exists {
			_ = tx.Rollback(ctx)
			return ErrParamsType
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
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// 缓存失效与重建
	_ = delCache(ctx, r.rdb, post.CacheHotDetailKey(postID, userID))
	_ = scanDelCache(ctx, r.rdb, post.CacheHotCommentsPattern(postID, userID), 100)
	if parentID != nil && strings.TrimSpace(*parentID) != "" {
		_ = scanDelCache(ctx, r.rdb, post.CacheCommentHotRepliesPattern(*parentID, userID), 100)
	}
	_ = scanDelCache(ctx, r.rdb, post.CacheHotListPattern(userID), 100)
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
	key := post.CacheHotCommentsKey(postID, userID, page, pageSize)
	if strings.ToLower(strings.TrimSpace(strategy)) == "hot" {
		var cached listCache
		if ok, _ := getCacheJSON(ctx, r.rdb, key, &cached); ok {
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
		payload := listCache{Rows: res, HasMore: hasMore}
		_ = setCacheJSON(ctx, r.rdb, key, payload, time.Duration(constant.POST_HOT_COMMENTS_TTL)*time.Second)
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
