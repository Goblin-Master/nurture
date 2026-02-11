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
		}
	})

	routeManager.RegisterUserRoutes(func(rg *gin.RouterGroup) {
		userHandler := handler.NewUserHandler()
		rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
		rg.POST("/register", middleware.BindJsonMiddleware[dto.RegisterReq], userHandler.Register)
		rg.POST("/code/login", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetLoginCode)
		rg.POST("/code/register", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetRegisterCode)
		rg.POST("/code/reset", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetResetCode)
		rg.POST("/resetPassword", middleware.BindJsonMiddleware[dto.ResetPasswordReq], userHandler.ResetPassword)
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
	})

	// 帖子模块路由
	routeManager.RegisterPostRoutes(func(rg *gin.RouterGroup) {
		postHandler := handler.NewPostHandler()
		// 首页帖子列表（公开）
		rg.GET("/list",
			middleware.BindQueryMiddleware[dto.PostListReq],
			postHandler.ListPosts,
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
		// 发布草稿
		rg.POST("/:post_id/publish",
			middleware.Authentication(jwtx.COMMON_USER),
			middleware.BindUriMiddleware[dto.PublishPostReq],
			postHandler.Publish,
		)
	})
}
