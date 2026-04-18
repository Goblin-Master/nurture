package cache

import (
	"fmt"
	"nurture/internal/constant"
)

func UserProfileKey(userID string) string {
	return fmt.Sprintf(constant.USER_PROFILE_KEY, userID)
}

func UserPartnerKey(userID string) string {
	return fmt.Sprintf(constant.USER_PARTNER_KEY, userID)
}

func UserFollowingKey(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, size)
}

func UserFollowersKey(userID string, page, size int) string {
	return fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, size)
}

