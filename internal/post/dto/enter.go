package dto

import "encoding/json"

type (
	PostDetailReq struct {
		PostID string `uri:"post_id" binding:"required"`
	}
	PostDetail struct {
		PostID         string          `json:"post_id"`
		AuthorID       string          `json:"author_id"`
		AuthorName     string          `json:"author_name"`
		AuthorAvatar   string          `json:"author_avatar"`
		AuthorProvince string          `json:"author_province"`
		AuthorCity     string          `json:"author_city"`
		Title          string          `json:"title"`
		Content        json.RawMessage `json:"content"`
		Status         string          `json:"status"`
		LikeCount      int32           `json:"like_count"`
		DislikeCount   int32           `json:"dislike_count"`
		CollectCount   int32           `json:"collect_count"`
		CommentCount   int32           `json:"comment_count"`
		Ctime          int64           `json:"ctime"`
		Utime          int64           `json:"utime"`
		Tags           []string        `json:"tags"`
		IsLike         bool            `json:"is_like"`
		IsDislike      bool            `json:"is_dislike"`
		IsCollect      bool            `json:"is_collect"`
		IsFollow       bool            `json:"is_follow"`
		BabyAgeYear    int             `json:"baby_age_year"`
		BabyAgeMonth   int             `json:"baby_age_month"`
		BabyAgeText    string          `json:"baby_age_text"`
	}
	PostDetailResp struct {
		Post PostDetail `json:"post"`
	}
)

type (
	CreatePostReq struct {
		Title   string          `json:"title" binding:"required"`
		Content json.RawMessage `json:"content" binding:"required"`
		Status  string          `json:"status"`
		TagIDs  []string        `json:"tag_ids"`
	}
	CreatePostResp struct {
		PostID  string `json:"post_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
)

type (
	UpdateDraftReq struct {
		Title   string          `json:"title" binding:"required"`
		Content json.RawMessage `json:"content" binding:"required"`
		TagIDs  []string        `json:"tag_ids"`
	}
	UpdateDraftResp struct {
		PostID  string `json:"post_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
)

type (
	PublishPostReq struct {
		PostID string `uri:"post_id" binding:"required"`
	}
	PublishPostResp struct {
		PostID  string `json:"post_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
)

type (
	CreateCommentReq struct {
		ParentID string `json:"parent_id"`
		Content  string `json:"content" binding:"required"`
	}
	CreateCommentResp struct {
		CommentID string `json:"comment_id"`
		Message   string `json:"message"`
	}
)

type (
	CommentListReq struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Strategy string `form:"strategy"`
	}
	CommentItem struct {
		CommentID  string `json:"comment_id"`
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		Avatar     string `json:"avatar"`
		Content    string `json:"content"`
		LikeCount  int32  `json:"like_count"`
		ReplyCount int32  `json:"reply_count"`
		Ctime      int64  `json:"ctime"`
		Utime      int64  `json:"utime"`
		HasLiked   bool   `json:"has_liked"`
	}
	CommentListResp struct {
		Items    []CommentItem `json:"items"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
		HasMore  bool          `json:"has_more"`
	}
)

type (
	CommentRepliesReq struct {
		PostID    string `uri:"post_id" binding:"required"`
		CommentID string `uri:"comment_id" binding:"required"`
	}
	CommentDeleteReq struct {
		CommentID string `uri:"comment_id" binding:"required"`
	}
	CommentUpdateReq struct {
		CommentID string `uri:"comment_id" binding:"required"`
	}
)

type (
	AdminTagCreateReq struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	AdminTagCreateResp struct {
		TagID       string `json:"tag_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	AdminTagDeleteUri struct {
		TagID string `uri:"tag_id" binding:"required"`
	}
	TagListReq struct {
		Keyword  string `form:"keyword"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	TagItem struct {
		TagID       string `json:"tag_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	TagListResp struct {
		Items    []TagItem `json:"items"`
		Page     int       `json:"page"`
		PageSize int       `json:"page_size"`
		HasMore  bool      `json:"has_more"`
	}
)

type (
	UpdateCommentReq struct {
		Content string `json:"content" binding:"required"`
	}
)

type (
	CommentLikeReq struct {
		CommentID string `uri:"comment_id" binding:"required"`
	}
)
type (
	// 首页列表
	PostHomeListReq struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Strategy string `form:"strategy"` // hot/ctime/random/recommend
	}
	// 按标签列表
	PostTagListReq struct {
		TagID    string `uri:"tag_id" binding:"required"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Strategy string `form:"strategy"` // hot/ctime
	}
	// 搜索列表
	PostSearchListReq struct {
		Keyword  string `form:"keyword" binding:"required"`
		TagID    string `form:"tag_id"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Strategy string `form:"strategy"` // hot/ctime
	}
	// 我的帖子/草稿列表
	PostMyListReq struct {
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
		Strategy string `form:"strategy"` // posts: hot/ctime; drafts: ctime only
	}
	PostItem struct {
		PostID         string          `json:"post_id"`
		AuthorID       string          `json:"author_id"`
		AuthorName     string          `json:"author_name"`
		AuthorAvatar   string          `json:"author_avatar"`
		AuthorProvince string          `json:"author_province"`
		AuthorCity     string          `json:"author_city"`
		Title          string          `json:"title"`
		Content        json.RawMessage `json:"content"`
		Status         string          `json:"status"`
		LikeCount      int32           `json:"like_count"`
		DislikeCount   int32           `json:"dislike_count"`
		CollectCount   int32           `json:"collect_count"`
		CommentCount   int32           `json:"comment_count"`
		Ctime          int64           `json:"ctime"`
		Utime          int64           `json:"utime"`
		Tags           []string        `json:"tags"`
		IsLike         bool            `json:"is_like"`
		IsDislike      bool            `json:"is_dislike"`
		IsCollect      bool            `json:"is_collect"`
		BabyAgeYear    int             `json:"baby_age_year"`
		BabyAgeMonth   int             `json:"baby_age_month"`
		BabyAgeText    string          `json:"baby_age_text"`
	}
	PostListResp struct {
		Items    []PostItem `json:"items"`
		Page     int        `json:"page"`
		PageSize int        `json:"page_size"`
		HasMore  bool       `json:"has_more"`
	}
)
type (
	CollectResp struct {
		CollectionID string `json:"collection_id"`
		Message      string `json:"message"`
	}
)
