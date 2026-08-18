package chat

import (
	"nurture/internal/dto"
	handlerchat "nurture/internal/handler/chat"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	chatHandler := handlerchat.NewHandler()
	registerGroupRoutes(rg.Group("/groups", middleware.Authentication(jwtx.COMMON_USER)), chatHandler)
	registerSessionRoutes(rg, chatHandler)
}

func registerGroupRoutes(groups *gin.RouterGroup, chatHandler *handlerchat.Handler) {
	groups.POST("",
		middleware.RateLimitUser("chat:groups:create", 10, time.Minute),
		middleware.BindJsonMiddleware[dto.CreateChatGroupReq],
		chatHandler.CreateGroup,
	)
	groups.GET("/discover",
		middleware.RateLimitUser("chat:groups:discover", 120, time.Minute),
		middleware.BindQueryMiddleware[dto.ChatGroupDiscoverReq],
		chatHandler.DiscoverGroups,
	)
	groups.GET("/search",
		middleware.RateLimitUser("chat:groups:search", 120, time.Minute),
		middleware.BindQueryMiddleware[dto.ChatGroupSearchReq],
		chatHandler.SearchGroups,
	)
	groups.GET("/mine",
		middleware.RateLimitUser("chat:groups:mine", 120, time.Minute),
		chatHandler.ListMyGroups,
	)
	groups.GET("/:group_id/profile",
		middleware.RateLimitUser("chat:groups:profile", 120, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		chatHandler.GroupProfile,
	)
	groups.POST("/:group_id/join",
		middleware.RateLimitUser("chat:groups:join", 30, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		chatHandler.JoinGroup,
	)
	groups.POST("/:group_id/leave",
		middleware.RateLimitUser("chat:groups:leave", 30, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		chatHandler.LeaveGroup,
	)
	groups.POST("/:group_id/transfer",
		middleware.RateLimitUser("chat:groups:transfer", 10, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		middleware.BindJsonMiddleware[dto.ChatGroupTransferReq],
		chatHandler.TransferOwner,
	)
	groups.POST("/:group_id/dissolve",
		middleware.RateLimitUser("chat:groups:dissolve", 10, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		chatHandler.DissolveGroup,
	)
	groups.POST("/:group_id/seen",
		middleware.RateLimitUser("chat:groups:seen", 60, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		chatHandler.MarkSeen,
	)
	groups.GET("/:group_id/members",
		middleware.RateLimitUser("chat:groups:members", 120, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		middleware.BindQueryMiddleware[dto.ChatGroupMemberListReq],
		chatHandler.ListMembers,
	)
	groups.GET("/:group_id/messages",
		middleware.RateLimitUser("chat:groups:messages", 120, time.Minute),
		middleware.BindUriMiddleware[dto.ChatGroupIDUri],
		middleware.BindQueryMiddleware[dto.ChatGroupMessageListReq],
		chatHandler.ListMessages,
	)
}

func registerSessionRoutes(sessions *gin.RouterGroup, chatHandler *handlerchat.Handler) {
	sessions.GET("/direct", chatHandler.OpenDirectSession)
	sessions.GET("/groups/session", chatHandler.OpenGroupSession)
}
