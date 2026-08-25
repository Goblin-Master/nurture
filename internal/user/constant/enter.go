package constant

import "time"

const (
	LoginWithAccount = "account"
	LoginWithEmail   = "email"
)

const (
	LoginCodeKey       = "login_code:%s"
	ResetPwdCodeKey    = "reset_pwd_code:%s"
	RegisterCodeKey    = "register_code:%s"
	RegisterSMSCodeKey = "register_sms_code:%s"
	BindPhoneCodeKey   = "bind_phone_code:%s"
	BindEmailCodeKey   = "bind_email_code:%s"
	RebindPhoneCodeKey = "rebind_phone_code:%s"
	RebindEmailCodeKey = "rebind_email_code:%s"
)

const (
	ProfileKey   = "user:profile:%s"
	PartnerKey   = "user:partner:%s"
	FollowingKey = "user:following:%s:%d:%d"
	FollowersKey = "user:followers:%s:%d:%d"
	TagPrefKey   = "user:tag_pref:%s"
)

const (
	ProfileTTL = 10 * 60
	PartnerTTL = 24 * 3600
	ListTTL    = 5 * 60
	TagPrefTTL = 30 * 24 * 3600
)

const (
	OutboxBatchSize      = 50
	OutboxMaxAttempts    = 8
	OutboxPollInterval   = time.Second
	OutboxRetryBaseDelay = time.Second
	OutboxClaimTimeout   = time.Minute
)
