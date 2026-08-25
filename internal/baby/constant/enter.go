package constant

import "time"

const (
	FeedBreastMilk = "breast_milk"
	FeedFormula    = "formula"
	FeedSolid      = "solid"
)

const (
	InfoKey         = "baby:info:%s:%s"
	VaccineListKey  = "baby:vaccine:list:%s"
	LatestGrowthKey = "baby:growth:latest:%s"
	InfoTTL         = 30 * 60
	VaccineListTTL  = 30 * 60
	LatestGrowthTTL = 5 * 60
)

const (
	UserEventExchange          = "user.event"
	UserPartnerBoundRoutingKey = "partner.bound"
	PartnerBoundQueue          = "baby.partner.bound"
	PartnerBoundConsumer       = "baby.partner.bound"
)

const (
	ConsumerRetryDelay  = 2 * time.Second
	ConsumerMaxAttempts = 3
)
