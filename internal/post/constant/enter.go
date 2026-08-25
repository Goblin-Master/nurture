package constant

const (
	RecommendCollection = "post_recommend"
)

const (
	HotListKey           = "post:list:hot:%d:%d"
	HotListByTagKey      = "post:list:hot:tag:%s:%d:%d"
	HotDetailKey         = "post:detail:hot:%s"
	HotCommentsKey       = "post:comments:hot:%s:%s:%d:%d"
	CommentHotRepliesKey = "comment:replies:hot:%s:%s:%d:%d"
	UserTagPrefKey       = "user:tag_pref:%s"
)

const (
	HotListTTL           = 2 * 60
	HotDetailTTL         = 10 * 60
	HotCommentsTTL       = 2 * 60
	CommentHotRepliesTTL = 2 * 60
	UserTagPrefTTL       = 30 * 24 * 3600
)
