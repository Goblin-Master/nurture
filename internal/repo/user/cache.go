package user

import (
	"fmt"
	"nurture/internal/constant"
)

func CacheProfileKey(userID string) string {
	return fmt.Sprintf(constant.USER_PROFILE_KEY, userID)
}

func CachePartnerKey(userID string) string {
	return fmt.Sprintf(constant.USER_PARTNER_KEY, userID)
}

func CacheFollowingKey(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, size)
}

func CacheFollowersKey(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, size)
}

func CacheFollowingPattern(userID string) string {
	return fmt.Sprintf("user:following:%s:*", userID)
}

func CacheFollowersPattern(userID string) string {
	return fmt.Sprintf("user:followers:%s:*", userID)
}

func CacheTagPrefKey(userID string) string {
	return fmt.Sprintf(constant.USER_TAG_PREF_KEY, userID)
}
