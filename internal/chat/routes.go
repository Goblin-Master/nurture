package chat

import (
	"nurture/internal/chat/constant"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(chatAPI *gin.RouterGroup, ws *gin.RouterGroup) {
	groups := chatAPI.Group("/groups", m.authUser)
	{
		groups.POST("",
			m.rateLimitUser(constant.RateLimitGroupCreate, constant.RateLimitGroupCreateLimit, constant.RateLimitHTTPWindow),
			m.handler.CreateGroup,
		)
		groups.GET("/discover",
			m.rateLimitUser(constant.RateLimitGroupDiscover, constant.RateLimitGroupDiscoverLimit, constant.RateLimitHTTPWindow),
			m.handler.DiscoverGroups,
		)
		groups.GET("/search",
			m.rateLimitUser(constant.RateLimitGroupSearch, constant.RateLimitGroupSearchLimit, constant.RateLimitHTTPWindow),
			m.handler.SearchGroups,
		)
		groups.GET("/mine",
			m.rateLimitUser(constant.RateLimitGroupMine, constant.RateLimitGroupMineLimit, constant.RateLimitHTTPWindow),
			m.handler.ListMyGroups,
		)
		groups.GET("/:group_id/profile",
			m.rateLimitUser(constant.RateLimitGroupProfile, constant.RateLimitGroupProfileLimit, constant.RateLimitHTTPWindow),
			m.handler.GroupProfile,
		)
		groups.POST("/:group_id/join",
			m.rateLimitUser(constant.RateLimitGroupJoin, constant.RateLimitGroupJoinLimit, constant.RateLimitHTTPWindow),
			m.handler.JoinGroup,
		)
		groups.POST("/:group_id/leave",
			m.rateLimitUser(constant.RateLimitGroupLeave, constant.RateLimitGroupLeaveLimit, constant.RateLimitHTTPWindow),
			m.handler.LeaveGroup,
		)
		groups.POST("/:group_id/transfer",
			m.rateLimitUser(constant.RateLimitGroupTransfer, constant.RateLimitGroupTransferLimit, constant.RateLimitHTTPWindow),
			m.handler.TransferOwner,
		)
		groups.POST("/:group_id/dissolve",
			m.rateLimitUser(constant.RateLimitGroupDissolve, constant.RateLimitGroupDissolveLimit, constant.RateLimitHTTPWindow),
			m.handler.DissolveGroup,
		)
		groups.POST("/:group_id/seen",
			m.rateLimitUser(constant.RateLimitGroupSeen, constant.RateLimitGroupSeenLimit, constant.RateLimitHTTPWindow),
			m.handler.MarkSeen,
		)
		groups.GET("/:group_id/members",
			m.rateLimitUser(constant.RateLimitGroupMembers, constant.RateLimitGroupMembersLimit, constant.RateLimitHTTPWindow),
			m.handler.ListMembers,
		)
		groups.GET("/:group_id/messages",
			m.rateLimitUser(constant.RateLimitGroupMessages, constant.RateLimitGroupMessagesLimit, constant.RateLimitHTTPWindow),
			m.handler.ListMessages,
		)
	}
	ws.GET("/chat", m.handler.ConnectDirect)
	ws.GET("/group", m.handler.ConnectGroup)
}
