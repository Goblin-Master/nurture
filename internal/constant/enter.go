package constant

// 所有常量文件读取位置
const (
	TOKEN_USER_ID         = "UserID"
	TOKEN_ROLE            = "Role"
	LOGIN_WITH_ACCOUNT    = "account"
	LOGIN_WITH_EMAIL      = "email"
	DEFAULT_NODE_ID       = 1
	FILE_MAX_SIZE         = 1024 * 1024 * 10
	LOGIN_CODE_KEY        = "login_code:%s"
	RESET_PWD_CODE_KEY    = "reset_pwd_code:%s"
	REGISTER_CODE_KEY     = "register_code:%s"
	REGISTER_SMS_CODE_KEY = "register_sms_code:%s"
	BIND_PHONE_CODE_KEY   = "bind_phone_code:%s"
	BIND_EMAIL_CODE_KEY   = "bind_email_code:%s"
	REBIND_PHONE_CODE_KEY = "rebind_phone_code:%s"
	REBIND_EMAIL_CODE_KEY = "rebind_email_code:%s"

	// 知识空间类型
	SPACE_TYPE_PRIVATE = "private"
	SPACE_TYPE_PUBLIC  = "public"

	// CollectionName 模板
	COLLECTION_USER_PREFIX = "knowledge_user_%s" //user_id
	COLLECTION_PUBLIC      = "knowledge_public"

	// Redis Key
	CHAT_HISTORY_KEY = "chat:history:%s:%s" // user_id:session_id

	// 对话历史
	AI_CONTEXT_MESSAGES = 6             // AI 上下文：最近 3 轮问答（6 条消息）
	HISTORY_TTL         = 7 * 24 * 3600 // 7 天（秒）

	// 用户缓存 Key
	USER_PROFILE_KEY   = "user:profile:%s"
	USER_PARTNER_KEY   = "user:partner:%s"
	USER_FOLLOWING_KEY = "user:following:%s:%d:%d"
	USER_FOLLOWERS_KEY = "user:followers:%s:%d:%d"
	// 用户缓存 TTL（秒）
	USER_PROFILE_TTL = 10 * 60
	USER_PARTNER_TTL = 24 * 3600
	USER_LIST_TTL    = 5 * 60

	// 帖子缓存 Key
	POST_HOT_LIST_KEY       = "post:list:hot:%d:%d"
	POST_HOT_LIST_BY_TAG    = "post:list:hot:tag:%s:%d:%d"
	POST_HOT_DETAIL_KEY     = "post:detail:hot:%s"
	POST_HOT_COMMENTS_KEY   = "post:comments:hot:%s:%s:%d:%d"
	COMMENT_HOT_REPLIES_KEY = "comment:replies:hot:%s:%s:%d:%d"
	// 帖子缓存 TTL（秒）
	POST_HOT_LIST_TTL       = 2 * 60
	POST_HOT_DETAIL_TTL     = 10 * 60
	POST_HOT_COMMENTS_TTL   = 2 * 60
	COMMENT_HOT_REPLIES_TTL = 2 * 60

	// 宝宝缓存 Key
	BABY_INFO_KEY          = "baby:info:%s:%s"
	BABY_VACCINE_LIST_KEY  = "baby:vaccine:list:%s"
	BABY_LATEST_GROWTH_KEY = "baby:growth:latest:%s"
	// 宝宝缓存 TTL（秒）
	BABY_INFO_TTL          = 30 * 60
	BABY_VACCINE_LIST_TTL  = 30 * 60
	BABY_LATEST_GROWTH_TTL = 5 * 60

	// SSE 事件类型
	SSE_TYPE_CONTENT = "content"
	SSE_TYPE_ERROR   = "error"
	SSE_TYPE_DONE    = "done"

	// 喂养方式
	FEED_BREAST_MILK = "breast_milk"
	FEED_FORMULA     = "formula"
	FEED_SOLID       = "solid"
)
