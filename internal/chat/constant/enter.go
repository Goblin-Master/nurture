package constant

import "time"

const (
	ChannelDirect = "direct"
	ChannelGroup  = "group"
)

const (
	MessageTypeText   = "text"
	MessageTypeImage  = "image"
	MessageTypeSystem = "system"
)

const (
	WSWriteWait      = 10 * time.Second
	WSPongWait       = 60 * time.Second
	WSPingPeriod     = (WSPongWait * 9) / 10
	WSMaxMessageSize = 4096
	WSSendBufferSize = 256
	WSDeliveredCache = 1024
)

const (
	OutboxBatchSize      = 50
	OutboxMaxAttempts    = 8
	OutboxPollInterval   = time.Second
	OutboxRetryBaseDelay = time.Second
	OutboxClaimTimeout   = time.Minute
)

const (
	ConsumerRetryDelay  = 2 * time.Second
	ConsumerMaxAttempts = 3
)

const (
	RateLimitGroupCreate    = "chat:groups:create"
	RateLimitGroupDiscover  = "chat:groups:discover"
	RateLimitGroupSearch    = "chat:groups:search"
	RateLimitGroupMine      = "chat:groups:mine"
	RateLimitGroupProfile   = "chat:groups:profile"
	RateLimitGroupJoin      = "chat:groups:join"
	RateLimitGroupLeave     = "chat:groups:leave"
	RateLimitGroupTransfer  = "chat:groups:transfer"
	RateLimitGroupDissolve  = "chat:groups:dissolve"
	RateLimitGroupSeen      = "chat:groups:seen"
	RateLimitGroupMembers   = "chat:groups:members"
	RateLimitGroupMessages  = "chat:groups:messages"
	RateLimitDirectMessages = "chat:direct:messages"
	RateLimitDirectSeen     = "chat:direct:seen"
)

const (
	RateLimitGroupCreateLimit    int64 = 10
	RateLimitGroupDiscoverLimit  int64 = 120
	RateLimitGroupSearchLimit    int64 = 120
	RateLimitGroupMineLimit      int64 = 120
	RateLimitGroupProfileLimit   int64 = 120
	RateLimitGroupJoinLimit      int64 = 30
	RateLimitGroupLeaveLimit     int64 = 30
	RateLimitGroupTransferLimit  int64 = 10
	RateLimitGroupDissolveLimit  int64 = 10
	RateLimitGroupSeenLimit      int64 = 60
	RateLimitGroupMembersLimit   int64 = 120
	RateLimitGroupMessagesLimit  int64 = 120
	RateLimitDirectMessagesLimit int64 = 120
	RateLimitDirectSeenLimit     int64 = 60
)

const (
	RateLimitHTTPWindow        = time.Minute
	RateLimitWSSendWindow      = time.Second
	RateLimitWSSendUserKey     = "rl:ws:send:user:%s"
	RateLimitWSSendDirectKey   = "rl:ws:send:direct:%s:%s"
	RateLimitWSSendGroupKey    = "rl:ws:send:group:%s:%s"
	RateLimitWSSendUserLimit   = 10
	RateLimitWSSendDirectLimit = 10
	RateLimitWSSendGroupLimit  = 5
)
