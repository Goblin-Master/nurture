package chat

import (
	"nurture/internal/chat/constant"
	"nurture/internal/chat/dto"
	"nurture/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(chatAPI *gin.RouterGroup, ws *gin.RouterGroup) {
	groups := chatAPI.Group("/groups", m.authUser)
	{
		groups.POST("",
			m.rateLimitUser(constant.RateLimitGroupCreate, constant.RateLimitGroupCreateLimit, constant.RateLimitHTTPWindow),
			middleware.BindJsonMiddleware[dto.CreateChatGroupReq],
			m.handler.CreateGroup,
		)
		groups.GET("/discover",
			m.rateLimitUser(constant.RateLimitGroupDiscover, constant.RateLimitGroupDiscoverLimit, constant.RateLimitHTTPWindow),
			middleware.BindQueryMiddleware[dto.ChatGroupDiscoverReq],
			m.handler.DiscoverGroups,
		)
		groups.GET("/search",
			m.rateLimitUser(constant.RateLimitGroupSearch, constant.RateLimitGroupSearchLimit, constant.RateLimitHTTPWindow),
			middleware.BindQueryMiddleware[dto.ChatGroupSearchReq],
			m.handler.SearchGroups,
		)
		groups.GET("/mine",
			m.rateLimitUser(constant.RateLimitGroupMine, constant.RateLimitGroupMineLimit, constant.RateLimitHTTPWindow),
			m.handler.ListMyGroups,
		)
		groups.GET("/:group_id/profile",
			m.rateLimitUser(constant.RateLimitGroupProfile, constant.RateLimitGroupProfileLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			m.handler.GroupProfile,
		)
		groups.POST("/:group_id/join",
			m.rateLimitUser(constant.RateLimitGroupJoin, constant.RateLimitGroupJoinLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			m.handler.JoinGroup,
		)
		groups.POST("/:group_id/leave",
			m.rateLimitUser(constant.RateLimitGroupLeave, constant.RateLimitGroupLeaveLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			m.handler.LeaveGroup,
		)
		groups.POST("/:group_id/transfer",
			m.rateLimitUser(constant.RateLimitGroupTransfer, constant.RateLimitGroupTransferLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			middleware.BindJsonMiddleware[dto.ChatGroupTransferReq],
			m.handler.TransferOwner,
		)
		groups.POST("/:group_id/dissolve",
			m.rateLimitUser(constant.RateLimitGroupDissolve, constant.RateLimitGroupDissolveLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			m.handler.DissolveGroup,
		)
		groups.POST("/:group_id/seen",
			m.rateLimitUser(constant.RateLimitGroupSeen, constant.RateLimitGroupSeenLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			m.handler.MarkSeen,
		)
		groups.GET("/:group_id/members",
			m.rateLimitUser(constant.RateLimitGroupMembers, constant.RateLimitGroupMembersLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			middleware.BindQueryMiddleware[dto.ChatGroupMemberListReq],
			m.handler.ListMembers,
		)
		groups.GET("/:group_id/messages",
			m.rateLimitUser(constant.RateLimitGroupMessages, constant.RateLimitGroupMessagesLimit, constant.RateLimitHTTPWindow),
			middleware.BindUriMiddleware[dto.ChatGroupIDUri],
			middleware.BindQueryMiddleware[dto.ChatGroupMessageListReq],
			m.handler.ListMessages,
		)
	}
	ws.GET("/chat", m.handler.ConnectDirect)
	ws.GET("/group", m.handler.ConnectGroup)
}
