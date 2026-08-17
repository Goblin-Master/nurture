package post

import (
	"fmt"
	"nurture/internal/constant"
)

func CacheHotDetailKey(postID, userID string) string {
	return fmt.Sprintf(constant.POST_HOT_DETAIL_KEY, postID) + ":" + userID
}

func CacheHotListKey(userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_LIST_KEY, page, pageSize) + ":" + userID
}

func CacheHotListByTagKey(userID, tagID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_LIST_BY_TAG, tagID, page, pageSize) + ":" + userID
}

func CacheHotCommentsKey(postID, userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_COMMENTS_KEY, postID, userID, page, pageSize)
}

func CacheCommentHotRepliesKey(commentID, userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.COMMENT_HOT_REPLIES_KEY, commentID, userID, page, pageSize)
}

func CacheHotListPattern(userID string) string {
	return fmt.Sprintf("post:list:hot:*:*:%s", userID)
}

func CacheHotCommentsPattern(postID, userID string) string {
	return fmt.Sprintf("post:comments:hot:%s:%s:*:*", postID, userID)
}

func CacheCommentHotRepliesPattern(commentID, userID string) string {
	return fmt.Sprintf("comment:replies:hot:%s:%s:*:*", commentID, userID)
}

func CacheHotDetailPattern(postID string) string {
	return fmt.Sprintf("post:detail:hot:%s:*", postID)
}

func CacheHotCommentsAllUsersPattern(postID string) string {
	return fmt.Sprintf("post:comments:hot:%s:*", postID)
}

func CacheHotListAllPattern() string {
	return "post:list:hot:*:*:*"
}

func CacheHotListByTagAllPattern() string {
	return "post:list:hot:tag:*:*:*:*"
}
