package router

import (
	"fmt"
	"nurture/internal/config"
	"nurture/internal/dto"
	"nurture/internal/handler"
	manager "nurture/internal/manger"
	"nurture/internal/middleware"
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
}
