package router

import (
	"fmt"
	"nurture/internal/config"
	"nurture/internal/middleware"
	routerchat "nurture/internal/router/chat"

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
	api := r.Group("/api")

	registerCommonRoutes(api.Group("/common"))
	routerchat.RegisterRoutes(api.Group("/chat"))
	registerUserRoutes(api.Group("/user"))
	registerBabyRoutes(api.Group("/baby"))
	registerPostRoutes(api.Group("/post"))
	registerAdminRoutes(api.Group("/admin"))
}
