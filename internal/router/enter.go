package router

import (
	"fmt"
	"nurture/internal/config"
	"nurture/internal/dto"
	"nurture/internal/handler"
	manager "nurture/internal/manger"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// RunServer 启动服务器 路由层
func RunServer() {
	r, err := listen()
	if err != nil {
		panic(err.Error())
	}
	err = r.Run(fmt.Sprintf("%s:%d", config.Conf.App.Host, config.Conf.App.Port)) // 启动 Gin 服务器
	if err != nil {
		panic(err.Error())
	}
}

// listen 配置 Gin 服务器
func listen() (*gin.Engine, error) {
	r := gin.Default() // 创建默认的 Gin 引擎
	// 注册全局中间件（例如获取 Trace ID）
	manager.RequestGlobalMiddleware(r)
	// 创建 RouteManager 实例
	routeManager := manager.NewRouteManager(r)
	// 注册各业务路由组的具体路由
	registerRoutes(routeManager)
	return r, nil
}

// registerRoutes 注册各业务路由的具体处理函数
func registerRoutes(routeManager *manager.RouteManager) {

	routeManager.RegisterCommonRoutes(func(rg *gin.RouterGroup) {
		rg.GET("/ping", func(c *gin.Context) {
			response.Response(c, "pong", nil)
		})
		commonHandler := handler.NewCommonHandler()
		rg.POST("/file/upload", middleware.Authentication(jwtx.COMMON_USER), commonHandler.UploadFile)

		// AI 相关接口
		ai := rg.Group("/ai")
		{
			// 知识库上传
			ai.POST("/knowledge/upload",
				middleware.Authentication(jwtx.COMMON_USER),
				middleware.BindJsonMiddleware[dto.KnowledgeUploadReq],
				commonHandler.UploadKnowledge,
			)

			// 流式对话
			ai.POST("/chat/stream",
				middleware.Authentication(jwtx.COMMON_USER),
				middleware.BindJsonMiddleware[dto.ChatStreamReq],
				commonHandler.ChatStream,
			)

			// 获取对话历史
			ai.GET("/chat/history",
				middleware.Authentication(jwtx.COMMON_USER),
				middleware.BindQueryMiddleware[dto.ChatHistoryReq],
				commonHandler.GetChatHistory,
			)

			// 成长曲线分析
			ai.POST("/growth/analysis",
				middleware.Authentication(jwtx.COMMON_USER),
				middleware.BindJsonMiddleware[dto.GrowthAnalysisReq],
				commonHandler.GrowthAnalysis,
			)

			ai.POST("/report/growth",
				middleware.Authentication(jwtx.COMMON_USER),
				middleware.BindJsonMiddleware[dto.GrowthReportReq],
				commonHandler.GrowthReport,
			)
		}
	})

	routeManager.RegisterChatRoutes(func(rg *gin.RouterGroup) {
		chatHandler := handler.NewChatHandler()

		groups := rg.Group("/groups", middleware.Authentication(jwtx.COMMON_USER))
		{
			// 创建群聊
			groups.POST("",
				middleware.RateLimitUser("chat:groups:create", 10, time.Minute),
				middleware.BindJsonMiddleware[dto.CreateChatGroupReq],
				chatHandler.CreateGroup,
			)
			// 群聊列表
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
			// 获取我的群聊列表
			groups.GET("/mine",
				middleware.RateLimitUser("chat:groups:mine", 120, time.Minute),
				chatHandler.ListMyGroups,
			)
			// 群聊详情
			groups.GET("/:group_id/profile",
				middleware.RateLimitUser("chat:groups:profile", 120, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				chatHandler.GroupProfile,
			)
			// 加入群聊
			groups.POST("/:group_id/join",
				middleware.RateLimitUser("chat:groups:join", 30, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				chatHandler.JoinGroup,
			)
			// 退出群聊
			groups.POST("/:group_id/leave",
				middleware.RateLimitUser("chat:groups:leave", 30, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				chatHandler.LeaveGroup,
			)
			// 转移群主
			groups.POST("/:group_id/transfer",
				middleware.RateLimitUser("chat:groups:transfer", 10, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				middleware.BindJsonMiddleware[dto.ChatGroupTransferReq],
				chatHandler.TransferOwner,
			)
			// 解散群聊
			groups.POST("/:group_id/dissolve",
				middleware.RateLimitUser("chat:groups:dissolve", 10, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				chatHandler.DissolveGroup,
			)
			// 标记已读
			groups.POST("/:group_id/seen",
				middleware.RateLimitUser("chat:groups:seen", 60, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				chatHandler.MarkSeen,
			)
			// 获取群聊成员列表
			groups.GET("/:group_id/members",
				middleware.RateLimitUser("chat:groups:members", 120, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				middleware.BindQueryMiddleware[dto.ChatGroupMemberListReq],
				chatHandler.ListMembers,
			)
			// 获取群聊消息列表
			groups.GET("/:group_id/messages",
				middleware.RateLimitUser("chat:groups:messages", 120, time.Minute),
				middleware.BindUriMiddleware[dto.ChatGroupIDUri],
				middleware.BindQueryMiddleware[dto.ChatGroupMessageListReq],
				chatHandler.ListMessages,
			)
		}
	})

	routeManager.RegisterUserRoutes(func(rg *gin.RouterGroup) {
		userHandler := handler.NewUserHandler()
		// 登录
		rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
		// 注册
		rg.POST("/register", middleware.BindJsonMiddleware[dto.RegisterReq], userHandler.Register)
		// 手机号注册
		rg.POST("/register/sms", middleware.BindJsonMiddleware[dto.RegisterSMSReq], userHandler.RegisterSMS)
		// 登录验证码
		rg.POST("/code/login", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetLoginCode)
		// 注册验证码
		rg.POST("/code/register", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetRegisterCode)
		// 手机号注册验证码
		rg.POST("/code/register/sms", middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetRegisterSMSCode)
		// 重置密码验证码
		rg.POST("/code/reset", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetResetCode)
		// 重置密码
		rg.POST("/resetPassword", middleware.BindJsonMiddleware[dto.ResetPasswordReq], userHandler.ResetPassword)
		// 更新个人资料
		rg.PUT("/profile", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.UpdateUserAdditionReq], userHandler.UpdateProfile)
		// 更新头像
		rg.PUT("/avatar", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.UpdateAvatarReq], userHandler.UpdateAvatar)
		// 绑定手机号/邮箱（需要验证码）
		rg.POST("/code/bind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetBindPhoneCode)
		rg.POST("/bind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindPhoneReq], userHandler.BindPhone)
		rg.POST("/code/bind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetBindEmailCode)
		rg.POST("/bind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindEmailReq], userHandler.BindEmail)
		// 换绑手机号/邮箱（需要验证码）
		rg.POST("/code/rebind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetRebindPhoneCode)
		rg.POST("/rebind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindPhoneReq], userHandler.RebindPhone)
		rg.POST("/code/rebind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetRebindEmailCode)
		rg.POST("/rebind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindEmailReq], userHandler.RebindEmail)
		// 我的资料
		rg.GET("/me", middleware.Authentication(jwtx.COMMON_USER), userHandler.MyProfile)
		// 另一半关系
		// 绑定另一半
		rg.POST("/partner/bind", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.PartnerBindReq], userHandler.BindPartner)
		// 获取另一半
		rg.GET("/partner", middleware.Authentication(jwtx.COMMON_USER), userHandler.GetPartner)
		// 关注关系
		// 关注用户
		rg.POST("/follow/:target_user_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.FollowReq],
			userHandler.Follow,
		)
		// 取消关注
		rg.DELETE("/follow/:target_user_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.FollowReq],
			userHandler.Unfollow,
		)
		// 我的关注列表
		rg.GET("/following",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.FollowingListReq],
			userHandler.ListFollowing,
		)
		// 我的粉丝列表
		rg.GET("/followers",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.FollowersListReq],
			userHandler.ListFollowers,
		)
	})

	// 宝宝模块路由
	routeManager.RegisterBabyRoutes(func(rg *gin.RouterGroup) {
		babyHandler := handler.NewBabyHandler()
		// 切换宝宝：列表
		rg.GET("/changeBaby",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.ChangeBabyReq],
			babyHandler.ChangeBaby,
		)
		// 新增宝宝
		rg.POST("/newBaby",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.NewBabyReq],
			babyHandler.NewBaby,
		)
		// 新增成长记录
		rg.POST("/growthRecords",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.CreateGrowthReq],
			babyHandler.CreateGrowth,
		)
		// 指定日期成长记录
		rg.GET("/growthRecord",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.GrowthAtReq],
			babyHandler.GetGrowthAt,
		)
		// 成长曲线
		rg.GET("/growthCurve",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.GrowthCurveReq],
			babyHandler.GrowthCurve,
		)
		// 睡眠计时：开始
		rg.POST("/:baby_id/daily/sleep/start",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.SleepStartUri],
			babyHandler.SleepStart,
		)
		// 睡眠计时：结束
		rg.POST("/:baby_id/daily/sleep/stop",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.SleepStopUri],
			middleware.BindJsonMiddleware[dto.SleepStopReq],
			babyHandler.SleepStop,
		)
		// 睡眠计时：当前活动
		rg.GET("/:baby_id/daily/sleep/active",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.SleepActiveUri],
			babyHandler.SleepActive,
		)
		// 睡眠计时：按日期查询
		rg.GET("/:baby_id/daily/sleep/byDate",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.SleepListAtUri],
			middleware.BindQueryMiddleware[dto.SleepListAtReq],
			babyHandler.ListSleepAt,
		)
		// 喂养记录：创建
		rg.POST("/:baby_id/daily/feeding",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.FeedingCreateUri],
			middleware.BindJsonMiddleware[dto.FeedingCreateReq],
			babyHandler.CreateFeeding,
		)
		// 喂养记录：更新
		rg.PUT("/:baby_id/daily/feeding/:feeding_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.FeedingUpdateUri],
			middleware.BindJsonMiddleware[dto.FeedingUpdateReq],
			babyHandler.UpdateFeeding,
		)
		// 喂养记录：按日期查询
		rg.GET("/:baby_id/daily/feeding/byDate",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.FeedingListAtUri],
			middleware.BindQueryMiddleware[dto.FeedingListAtReq],
			babyHandler.ListFeedingAt,
		)
		// 换尿布记录：创建/更新
		rg.POST("/:baby_id/daily/diaper",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.DiaperCreateUri],
			middleware.BindJsonMiddleware[dto.DiaperCreateReq],
			babyHandler.CreateDiaper,
		)
		rg.PUT("/:baby_id/daily/diaper/:diaper_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.DiaperUpdateUri],
			middleware.BindJsonMiddleware[dto.DiaperUpdateReq],
			babyHandler.UpdateDiaper,
		)
		// 换尿布记录：按日期查询（仅当日一条）
		rg.GET("/:baby_id/daily/diaper/byDate",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.DiaperGetAtUri],
			middleware.BindQueryMiddleware[dto.DiaperGetAtReq],
			babyHandler.GetDiaperAt,
		)
		// 换尿布记录：按日期查询（列表，升序）
		rg.GET("/:baby_id/daily/diaper/byDate/list",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.DiaperListAtUri],
			middleware.BindQueryMiddleware[dto.DiaperListAtReq],
			babyHandler.ListDiaperAt,
		)
		// 日统计：按日期查询
		rg.GET("/:baby_id/daily/stats/byDate",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.DailyStatsUri],
			middleware.BindQueryMiddleware[dto.DailyStatsReq],
			babyHandler.DailyStats,
		)
		// 宝宝详细页
		rg.GET("/profile",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.BabyProfileReq],
			babyHandler.GetProfile,
		)
		// 疫苗列表
		rg.GET("/vaccine/getVaccineList",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.GetVaccineListReq],
			babyHandler.GetVaccineList,
		)
		// 更新接种状态
		rg.PUT("/vaccine/changeStatus",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.ChangeVaccineStatusReq],
			babyHandler.ChangeVaccineStatus,
		)
		// 上传宝宝照片
		rg.POST("/photo/upload",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.UploadBabyPhotosReq],
			babyHandler.UploadBabyPhotos,
		)
		// 删除宝宝照片
		rg.DELETE("/photo/delete",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.DeleteBabyPhotosReq],
			babyHandler.DeleteBabyPhotos,
		)
		// 获取宝宝照片
		rg.GET("/photo/list",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.ListBabyPhotosReq],
			babyHandler.ListBabyPhotos,
		)
	})

	// 帖子模块路由
	routeManager.RegisterPostRoutes(func(rg *gin.RouterGroup) {
		postHandler := handler.NewPostHandler()
		// 首页帖子列表（公开）
		// 首页（简化筛选）
		rg.GET("",
			middleware.BindQueryMiddleware[dto.PostHomeListReq],
			postHandler.Home,
		)
		// 标签选择列表
		rg.GET("/tags",
			middleware.BindQueryMiddleware[dto.TagListReq],
			postHandler.ListTags,
		)
		// 按标签列表
		rg.GET("/tag/:tag_id",
			middleware.BindUriMiddleware[dto.PostTagListReq],
			middleware.BindQueryMiddleware[dto.PostTagListReq],
			postHandler.ListByTag,
		)
		// 搜索列表
		rg.GET("/search",
			middleware.BindQueryMiddleware[dto.PostSearchListReq],
			postHandler.Search,
		)
		// 帖子详情
		rg.GET("/:post_id",
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.GetDetail,
		)
		// 新增帖子/草稿
		rg.POST("/newPost",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindJsonMiddleware[dto.CreatePostReq],
			postHandler.NewPost,
		)
		// 我的发布列表
		rg.GET("/mine",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.PostMyListReq],
			postHandler.ListMyPosts,
		)
		// 关注用户的帖子列表
		rg.GET("/following",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.PostMyListReq],
			postHandler.Following,
		)
		// 我的草稿列表
		rg.GET("/mine/drafts",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.PostMyListReq],
			postHandler.ListMyDrafts,
		)
		// 我的里程碑列表（大事记）
		rg.GET("/mine/milestone",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.PostMyListReq],
			postHandler.ListMyMilestones,
		)
		// 发布草稿
		rg.POST("/:post_id/publish",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PublishPostReq],
			postHandler.Publish,
		)
		// 更新草稿
		rg.PUT("/:post_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			middleware.BindJsonMiddleware[dto.UpdateDraftReq],
			postHandler.UpdateDraft,
		)
		// 删除草稿
		rg.DELETE("/:post_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.DeleteDraft,
		)
		// 删除帖子（已发布/里程碑）
		rg.DELETE("/:post_id/delete",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.DeletePost,
		)
		// 点赞帖子
		rg.POST("/:post_id/like",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.LikePost,
		)
		// 取消点赞帖子
		rg.DELETE("/:post_id/like",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.UnlikePost,
		)
		// 收藏帖子
		rg.POST("/:post_id/collect",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.CollectPost,
		)
		// 取消收藏帖子
		rg.DELETE("/:post_id/collect",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PostDetailReq],
			postHandler.UncollectPost,
		)
		// 我的收藏列表
		rg.GET("/mine/collections",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindQueryMiddleware[dto.PostMyListReq],
			postHandler.ListMyCollections,
		)
		// 创建评论
		rg.POST("/:post_id/comments",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PublishPostReq],
			middleware.BindJsonMiddleware[dto.CreateCommentReq],
			postHandler.CreateComment,
		)
		// 一级评论列表
		rg.GET("/:post_id/comments",
			middleware.BindUriMiddleware[dto.PostDetailReq],
			middleware.BindQueryMiddleware[dto.CommentListReq],
			postHandler.ListComments,
		)
		// 子评论列表
		rg.GET("/:post_id/comments/:comment_id/replies",
			middleware.BindUriMiddleware[dto.CommentRepliesReq],
			middleware.BindQueryMiddleware[dto.CommentListReq],
			postHandler.ListReplies,
		)
		// 删除评论（软删除，仅作者）
		rg.DELETE("/comments/:comment_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.CommentDeleteReq],
			postHandler.DeleteComment,
		)
		// 修改评论（仅作者）
		rg.PUT("/comments/:comment_id",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.CommentUpdateReq],
			middleware.BindJsonMiddleware[dto.UpdateCommentReq],
			postHandler.UpdateComment,
		)
		// 点赞评论
		rg.POST("/comments/:comment_id/like",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.CommentLikeReq],
			postHandler.LikeComment,
		)
		// 取消点赞评论
		rg.DELETE("/comments/:comment_id/like",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.CommentLikeReq],
			postHandler.UnlikeComment,
		)
	})

	// 管理员模块路由
	routeManager.RegisterAdminRoutes(func(rg *gin.RouterGroup) {
		userHandler := handler.NewUserHandler()
		postHandler := handler.NewPostHandler()
		babyHandler := handler.NewBabyHandler()
		rg.GET("/users/list",
			middleware.Authentication(jwtx.ADMIN),
			middleware.BindQueryMiddleware[dto.AdminListUsersReq],
			userHandler.AdminListUsers,
		)
		rg.PUT("/users/:user_id/role/admin",
			middleware.Authentication(jwtx.ADMIN),
			middleware.BindUriMiddleware[dto.AdminPromoteUri],
			userHandler.AdminPromoteToAdmin,
		)
		// 标签管理
		rg.POST("/tags",
			middleware.Authentication(jwtx.ADMIN),
			middleware.BindJsonMiddleware[dto.AdminTagCreateReq],
			postHandler.AdminCreateTag,
		)
		rg.DELETE("/tags/:tag_id",
			middleware.Authentication(jwtx.ADMIN),
			middleware.BindUriMiddleware[dto.AdminTagDeleteUri],
			postHandler.AdminDeleteTag,
		)
		// 疫苗管理（仅创建）
		rg.POST("/vaccines",
			middleware.Authentication(jwtx.ADMIN),
			middleware.BindJsonMiddleware[dto.AdminCreateVaccineReq],
			babyHandler.AdminCreateVaccine,
		)
	})

	// WebSocket 路由
	routeManager.RegisterWSRoutes(func(rg *gin.RouterGroup) {
		wsHandler := handler.NewWebSocketHandler()
		rg.GET("/chat", wsHandler.Connect)
		rg.GET("/groups", wsHandler.ConnectGroups)
	})
}
