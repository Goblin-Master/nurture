package dto

type (
	PostDetailReq struct {
		PostID string `uri:"post_id" binding:"required"`
	}
	PostDetail struct {
		PostID         string   `json:"post_id"`
		AuthorID       string   `json:"author_id"`
		AuthorName     string   `json:"author_name"`
		AuthorAvatar   string   `json:"author_avatar"`
		AuthorProvince string   `json:"author_province"`
		AuthorCity     string   `json:"author_city"`
		Title          string   `json:"title"`
		Content        string   `json:"content"`
		Status         string   `json:"status"`
		LikeCount      int32    `json:"like_count"`
		DislikeCount   int32    `json:"dislike_count"`
		CollectCount   int32    `json:"collect_count"`
		CommentCount   int32    `json:"comment_count"`
		Cover          string   `json:"cover"`
		Ctime          int64    `json:"ctime"`
		Utime          int64    `json:"utime"`
		Tags           []string `json:"tags"`
		BabyAgeYear    int      `json:"baby_age_year"`
		BabyAgeMonth   int      `json:"baby_age_month"`
		BabyAgeText    string   `json:"baby_age_text"`
	}
	PostDetailResp struct {
		Post PostDetail `json:"post"`
	}
)

type (
	CreatePostReq struct {
		Title   string   `json:"title" binding:"required"`
		Content string   `json:"content" binding:"required"`
		Cover   string   `json:"cover"`
		Status  string   `json:"status"`
		TagIDs  []string `json:"tag_ids"`
	}
	CreatePostResp struct {
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
	PostListReq struct {
		Page       int    `form:"page"`
		PageSize   int    `form:"page_size"`
		Status     string `form:"status"`
		TagID      string `form:"tag_id"`
		AuthorID   string `form:"author_id"`
		OrderBy    string `form:"order_by"`
		Order      string `form:"order"`
		Keyword    string `form:"keyword"`
		Strategy   string `form:"strategy"`
		ExcludeIDs string `form:"exclude_ids"`
	}
	PostItem struct {
		PostID         string   `json:"post_id"`
		AuthorID       string   `json:"author_id"`
		AuthorName     string   `json:"author_name"`
		AuthorAvatar   string   `json:"author_avatar"`
		AuthorProvince string   `json:"author_province"`
		AuthorCity     string   `json:"author_city"`
		Title          string   `json:"title"`
		Content        string   `json:"content"`
		ContentPreview string   `json:"content_preview"`
		Status         string   `json:"status"`
		LikeCount      int32    `json:"like_count"`
		DislikeCount   int32    `json:"dislike_count"`
		CollectCount   int32    `json:"collect_count"`
		CommentCount   int32    `json:"comment_count"`
		Cover          string   `json:"cover"`
		Ctime          int64    `json:"ctime"`
		Utime          int64    `json:"utime"`
		Tags           []string `json:"tags"`
		BabyAgeYear    int      `json:"baby_age_year"`
		BabyAgeMonth   int      `json:"baby_age_month"`
		BabyAgeText    string   `json:"baby_age_text"`
	}
	PostListResp struct {
		Items    []PostItem `json:"items"`
		Page     int        `json:"page"`
		PageSize int        `json:"page_size"`
		HasMore  bool       `json:"has_more"`
	}
)
