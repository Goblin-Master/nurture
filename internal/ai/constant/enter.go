package constant

const (
	SpaceTypePrivate = "private"
	SpaceTypePublic  = "public"
)

const (
	CollectionUserPrefix = "knowledge_user_%s"
	CollectionPublic     = "knowledge_public"
)

const (
	ChatHistoryKey = "chat:history:%s:%s"
)

const (
	ContextMessages = 6
	HistoryTTL      = 7 * 24 * 3600
)

const (
	SSETypeContent = "content"
	SSETypeError   = "error"
	SSETypeDone    = "done"
)
