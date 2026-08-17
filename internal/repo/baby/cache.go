package baby

import (
	"fmt"
	"nurture/internal/constant"
)

func CacheInfoKey(babyID, userID string) string {
	return fmt.Sprintf(constant.BABY_INFO_KEY, babyID, userID)
}

func CacheVaccineListKey(babyID string) string {
	return fmt.Sprintf(constant.BABY_VACCINE_LIST_KEY, babyID)
}

func CacheLatestGrowthKey(babyID string) string {
	return fmt.Sprintf(constant.BABY_LATEST_GROWTH_KEY, babyID)
}

func CacheInfoPattern(babyID string) string {
	return fmt.Sprintf("baby:info:%s:*", babyID)
}
