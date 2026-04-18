package cache

import (
	"fmt"
	"nurture/internal/constant"
)

func BabyInfoKey(babyID, userID string) string {
	return fmt.Sprintf(constant.BABY_INFO_KEY, babyID, userID)
}

func BabyVaccineListKey(babyID string) string {
	return fmt.Sprintf(constant.BABY_VACCINE_LIST_KEY, babyID)
}

func BabyLatestGrowthKey(babyID string) string {
	return fmt.Sprintf(constant.BABY_LATEST_GROWTH_KEY, babyID)
}

func BabyInfoPattern(babyID string) string {
	return fmt.Sprintf("baby:info:%s:*", babyID)
}
