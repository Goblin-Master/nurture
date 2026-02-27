package constant

// 所有常量文件读取位置
const (
	TOKEN_USER_ID      = "UserID"
	TOKEN_ROLE         = "Role"
	LOGIN_WITH_ACCOUNT = "account"
	LOGIN_WITH_EMAIL   = "email"
	DEFAULT_NODE_ID    = 1
	FILE_MAX_SIZE      = 1024 * 1024 * 10
	LOGIN_CODE_KEY     = "login_code:%s"
	RESET_PWD_CODE_KEY = "reset_pwd_code:%s"
	REGISTER_CODE_KEY  = "register_code:%s"

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

	// SSE 事件类型
	SSE_TYPE_CONTENT = "content"
	SSE_TYPE_ERROR   = "error"
	SSE_TYPE_DONE    = "done"

	// 喂养方式
	FEED_BREAST_MILK = "breast_milk"
	FEED_FORMULA     = "formula"
	FEED_SOLID       = "solid"
)
