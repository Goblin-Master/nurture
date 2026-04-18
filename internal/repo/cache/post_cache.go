package cache

import (
	"fmt"
	"nurture/internal/constant"
)

func PostHotDetailKey(postID, userID string) string {
	return fmt.Sprintf(constant.POST_HOT_DETAIL_KEY, postID) + ":" + userID
}

func PostHotListKey(userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_LIST_KEY, page, pageSize) + ":" + userID
}

func PostHotListByTagKey(userID, tagID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_LIST_BY_TAG, tagID, page, pageSize) + ":" + userID
}

func PostHotCommentsKey(postID, userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.POST_HOT_COMMENTS_KEY, postID, userID, page, pageSize)
}

func CommentHotRepliesKey(commentID, userID string, page, pageSize int) string {
	return fmt.Sprintf(constant.COMMENT_HOT_REPLIES_KEY, commentID, userID, page, pageSize)
}

func PostHotListPattern(userID string) string {
	return fmt.Sprintf("post:list:hot:*:*:%s", userID)
}

func PostHotCommentsPattern(postID, userID string) string {
	return fmt.Sprintf("post:comments:hot:%s:%s:*:*", postID, userID)
}

func CommentHotRepliesPattern(commentID, userID string) string {
	return fmt.Sprintf("comment:replies:hot:%s:%s:*:*", commentID, userID)
}

