package router

import (
	"fmt"
	"nurture/internal/baby"
	babyrepo "nurture/internal/baby/repo"
	"nurture/internal/chat"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/post"
	"nurture/internal/user"

	"github.com/gin-gonic/gin"
)

// RunServer 启动服务器 路由层
func RunServer() {
	r, err := listen()
	if err != nil {
		panic(err.Error())
	}
	err = r.Run(fmt.Sprintf("%s:%d", config.Conf.App.Host, config.Conf.App.Port))
	if err != nil {
		panic(err.Error())
	}
}

// listen 配置 Gin 服务器
func listen() (*gin.Engine, error) {
	r := gin.Default()
	requestGlobalMiddleware(r)
	registerRoutes(r)
	return r, nil
}

// requestGlobalMiddleware 注册全局中间件，应用于所有路由
func requestGlobalMiddleware(r *gin.Engine) {
	r.Use(middleware.Cors())
}

// registerRoutes 注册各业务路由的具体处理函数
func registerRoutes(r *gin.Engine) {
	registerHealthRoutes(r)

	api := r.Group("/api")
	ws := r.Group("/ws")
	dbRequired := middleware.RequireDB()
	userModule := user.NewModule(user.Deps{
		DB:         global.DB,
		RDB:        global.RDB,
		Log:        global.Log,
		BabySyncer: babyrepo.NewBabyRepo(global.DB, global.RDB, global.Log),
	})
	babyModule := baby.NewModule(baby.Deps{
		DB:            global.DB,
		RDB:           global.RDB,
		Log:           global.Log,
		PartnerReader: userModule,
	})
	postModule := post.NewModule(post.Deps{
		DB:           global.DB,
		RDB:          global.RDB,
		Log:          global.Log,
		AI:           global.AIX,
		FollowReader: userModule,
	})

	registerCommonRoutes(api.Group("/common"))
	chat.NewModule(chat.Deps{
		DB:            global.DB,
		RDB:           global.RDB,
		RabbitMQ:      global.RMQ,
		Log:           global.Log,
		AuthUser:      middleware.Authentication(jwtx.COMMON_USER),
		RateLimitUser: middleware.RateLimitUser,
	}).RegisterRoutes(api.Group("/chat", dbRequired), ws.Group("", dbRequired))
	userModule.RegisterRoutes(api.Group("/user", dbRequired))
	babyModule.RegisterRoutes(api.Group("/baby", dbRequired))
	postModule.RegisterRoutes(api.Group("/post", dbRequired))
	admin := api.Group("/admin", dbRequired)
	userModule.RegisterAdminRoutes(admin)
	babyModule.RegisterAdminRoutes(admin)
	postModule.RegisterAdminRoutes(admin)
}
